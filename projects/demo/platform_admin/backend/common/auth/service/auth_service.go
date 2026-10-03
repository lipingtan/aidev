package service

import (
	"log"
	"sync"
	"time"

	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"
	"go-admin/common/auth/spi"
	"go-admin/common/auth/strategy"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
// blacklist 使用 spi.TokenBlacklistStore 接口，支持内存（单实例）或 Redis（多 Pod）实现
type AuthService struct {
	db             *gorm.DB
	cfg            *config.Config
	blacklist      spi.TokenBlacklistStore
	tokenRecorder  *TokenRecordService
}

// GetConfig 返回配置（供中间件等外部调用方使用）
func (s *AuthService) GetConfig() *config.Config {
	return s.cfg
}

// CheckTenantStatus 检查租户状态
func (s *AuthService) CheckTenantStatus(tenantID int64) (int, error) {
	var status int
	err := s.db.Table("admin_tenant").Where("id = ?", tenantID).Pluck("status", &status).Error
	return status, err
}

// AutoDegradeIfExpired 自动降级：如果租户 status=1 且 expired_at 已过期，CAS 更新为 READ_ONLY(2)
// 返回 true 表示已降级
func (s *AuthService) AutoDegradeIfExpired(tenantID int64) bool {
	result := s.db.Table("admin_tenant").
		Where("id = ? AND status = 1 AND expired_at IS NOT NULL AND expired_at < NOW()", tenantID).
		Update("status", 2)
	return result.RowsAffected > 0
}

// CheckUserStatus 检查用户状态
func (s *AuthService) CheckUserStatus(userID int64) (int, error) {
	var status int
	result := s.db.Table("admin_user").Where("id = ?", userID).Pluck("status", &status)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		// 用户不存在于 admin_user 表（可能是C端用户），返回 ErrRecordNotFound
		return 0, gorm.ErrRecordNotFound
	}
	return status, nil
}

// NewAuthService 构造认证服务
// blacklistStore 传 nil 时降级为本地内存黑名单（仅适用于单实例）
func NewAuthService(db *gorm.DB, cfg *config.Config, blacklistStore ...spi.TokenBlacklistStore) *AuthService {
	var bl spi.TokenBlacklistStore
	if len(blacklistStore) > 0 && blacklistStore[0] != nil {
		bl = blacklistStore[0]
	} else {
		bl = &localBlacklistFallback{}
	}
	tokenRecorder := NewTokenRecordService(db)
	tokenRecorder.SetBlacklist(bl) // 注入黑名单，使登录时旧 token 真正失效
	return &AuthService{
		db:            db,
		cfg:           cfg,
		blacklist:     bl,
		tokenRecorder: tokenRecorder,
	}
}

// localBlacklistFallback 单实例降级实现（Redis 不可用时使用）
type localBlacklistFallback struct {
	m sync.Map
}

func (l *localBlacklistFallback) Add(jti string, ttl time.Duration) error {
	l.m.Store(jti, time.Now().Add(ttl))
	return nil
}

func (l *localBlacklistFallback) Contains(jti string) bool {
	val, ok := l.m.Load(jti)
	if !ok {
		return false
	}
	if time.Now().After(val.(time.Time)) {
		l.m.Delete(jti)
		return false
	}
	return true
}

// GetUserIDByUsername 根据用户名获取用户 ID（用于登录日志写入，查不到返回 0）
func (s *AuthService) GetUserIDByUsername(username string) int64 {
	var user model.User
	if err := s.db.Select("id").Where("username = ?", username).First(&user).Error; err != nil {
		return 0
	}
	return user.ID
}

// Login 用户登录
func (s *AuthService) Login(username, password, ip, userAgent string) (*strategy.LoginResponse, error) {
	// 查询用户
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrInvalidCredentials, "用户名或密码错误")
		}
		return nil, err
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.NewAuthError(errors.ErrAccountDisabled, "账号已禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.NewAuthError(errors.ErrInvalidCredentials, "用户名或密码错误")
	}

	// 查询用户关联的租户
	var userTenants []model.UserTenant
	s.db.Where("user_id = ?", user.ID).Find(&userTenants)

	if len(userTenants) == 0 {
		return nil, errors.NewAuthError(errors.ErrNotAssociatedTenant, "用户未关联任何租户")
	}

	// 检查用户是否为平台超管（拥有 SUPER_ADMIN 角色）
	var superAdminCount int64
	s.db.Model(&model.UserRole{}).
		Joins("JOIN admin_role ON admin_role.id = admin_user_role.role_id").
		Where("admin_user_role.user_id = ? AND admin_role.role_code = ?", user.ID, s.cfg.Permission.SuperAdminRole).
		Count(&superAdminCount)
	isSuperAdmin := superAdminCount > 0

	var tenants []model.Tenant
	if isSuperAdmin {
		// 平台超管：加载所有可用租户
		s.db.Where("status = 1").Find(&tenants)
	} else {
		// 普通用户：仅加载已关联的租户
		tenantIDs := make([]int64, len(userTenants))
		for i, ut := range userTenants {
			tenantIDs[i] = ut.TenantID
		}
		s.db.Where("id IN ? AND status = 1", tenantIDs).Find(&tenants)
	}

	if len(tenants) == 0 {
		return nil, errors.NewAuthError(errors.ErrTenantDisabled, "所有关联租户均已禁用")
	}

	tenantInfos := make([]strategy.TenantInfo, len(tenants))
	for i, t := range tenants {
		tenantInfos[i] = strategy.TenantInfo{
			ID:   t.ID,
			Code: t.TenantCode,
			Name: t.Name,
		}
	}

	// 单租户优化：直接签发 access_token
	if len(tenantInfos) == 1 {
		platformToken, err := strategy.GeneratePlatformToken(s.cfg, user.ID, tenantInfos)
		if err != nil {
			return nil, err
		}

		// 查询用户在该租户下的角色
		roleIDs := s.getUserRoles(user.ID, tenantInfos[0].ID)

		// 计算 access token 过期时间（秒）：优先使用用户自定义，否则使用全局配置
		accessExpiresIn := int64(s.cfg.JWT.AccessTokenTTL.Seconds())
		if user.AccessTokenExpireMinutes != nil && *user.AccessTokenExpireMinutes > 0 {
			accessExpiresIn = int64(*user.AccessTokenExpireMinutes * 60)
		}

		accessToken, err := strategy.GenerateAccessToken(s.cfg, &strategy.AccessTokenOptions{
			UserID:       user.ID,
			TenantID:     tenantInfos[0].ID,
			Roles:        roleIDs,
			UserPool:     strategy.UserPoolAdmin,
			TokenVersion: 0,
			ExpiresIn:    accessExpiresIn,
		})
		if err != nil {
			return nil, err
		}

		// 计算 refresh token 过期时间（秒）：优先使用用户自定义，否则使用全局配置
		refreshExpiresIn := int64(s.cfg.JWT.RefreshTokenTTL.Seconds())
		if user.RefreshTokenExpireMinutes != nil && *user.RefreshTokenExpireMinutes > 0 {
			refreshExpiresIn = int64(*user.RefreshTokenExpireMinutes * 60)
		}

		refreshToken, err := strategy.GenerateRefreshToken(s.cfg, &strategy.RefreshTokenOptions{
			UserID:       user.ID,
			TenantID:     tenantInfos[0].ID,
			UserPool:     strategy.UserPoolAdmin,
			TokenVersion: 0,
			ExpiresIn:    refreshExpiresIn,
		})
		if err != nil {
			return nil, err
		}

		// 记录 Token 到审计日志
		if s.tokenRecorder != nil {
			accessClaims, _ := strategy.ParseAccessToken(s.cfg, accessToken)
			refreshClaims, _ := strategy.ParseRefreshToken(s.cfg, refreshToken)
			accessJTI := ""
			refreshJTI := ""
			if accessClaims != nil {
				accessJTI = accessClaims.ID
			}
			if refreshClaims != nil {
				refreshJTI = refreshClaims.ID
			}
			_ = s.tokenRecorder.RecordLoginFull(
				user.ID, user.Username, tenantInfos[0].ID,
				accessToken, accessJTI, time.Now().Add(time.Duration(accessExpiresIn)*time.Second),
				refreshToken, refreshJTI, time.Now().Add(time.Duration(refreshExpiresIn)*time.Second),
				model.TokenSourceLogin, ip, userAgent,
			)
		}

		return &strategy.LoginResponse{
			TokenType:     "access",
			Token:         accessToken,
			AccessToken:   accessToken,
			PlatformToken: platformToken,
			RefreshToken:  refreshToken,
			Tenants:       tenantInfos,
			ExpiresIn:     accessExpiresIn,
		}, nil
	}

	// 多租户：签发 platform_token
	platformToken, err := strategy.GeneratePlatformToken(s.cfg, user.ID, tenantInfos)
	if err != nil {
		return nil, err
	}

	return &strategy.LoginResponse{
		TokenType: "platform",
		Token:     platformToken,
		Tenants:   tenantInfos,
		ExpiresIn: int64(s.cfg.JWT.PlatformTokenTTL.Seconds()),
	}, nil
}

// SelectTenant 选择租户，签发 access_token
func (s *AuthService) SelectTenant(platformTokenStr string, tenantID int64, ip, userAgent string) (*strategy.SelectTenantResponse, error) {
	// 解析 platform_token
	claims, err := strategy.ParsePlatformToken(s.cfg, platformTokenStr)
	if err != nil {
		return nil, err
	}

	// 检查用户是否为平台超管（跳过租户关联校验）
	var superAdminCount int64
	s.db.Model(&model.UserRole{}).
		Joins("JOIN admin_role ON admin_role.id = admin_user_role.role_id").
		Where("admin_user_role.user_id = ? AND admin_role.role_code = ?", claims.UserID, s.cfg.Permission.SuperAdminRole).
		Count(&superAdminCount)

	if superAdminCount == 0 {
		// 普通用户：验证租户在 claims 的租户列表中
		found := false
		for _, t := range claims.Tenants {
			if t.ID == tenantID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.NewAuthError(errors.ErrNotAssociatedTenant, "用户未关联该租户")
		}
	}

	// 验证租户状态
	var tenant model.Tenant
	if err := s.db.Where("id = ? AND status = 1", tenantID).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewAuthError(errors.ErrTenantDisabled, "租户已禁用或不存在")
		}
		return nil, err
	}

	// 查询用户在该租户下的角色
	roleIDs := s.getUserRoles(claims.UserID, tenantID)

	// 查询用户以获取 per-user token expiration 设置
	var user model.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	// 计算 access token 过期时间（秒）：优先使用用户自定义，否则使用全局配置
	accessExpiresIn := int64(s.cfg.JWT.AccessTokenTTL.Seconds())
	if user.AccessTokenExpireMinutes != nil && *user.AccessTokenExpireMinutes > 0 {
		accessExpiresIn = int64(*user.AccessTokenExpireMinutes * 60)
	}

	// 签发 access_token
	accessToken, err := strategy.GenerateAccessToken(s.cfg, &strategy.AccessTokenOptions{
		UserID:       claims.UserID,
		TenantID:     tenantID,
		Roles:        roleIDs,
		UserPool:     strategy.UserPoolAdmin,
		TokenVersion: 0,
		ExpiresIn:    accessExpiresIn,
	})
	if err != nil {
		return nil, err
	}

	// 计算 refresh token 过期时间（秒）：优先使用用户自定义，否则使用全局配置
	refreshExpiresIn := int64(s.cfg.JWT.RefreshTokenTTL.Seconds())
	if user.RefreshTokenExpireMinutes != nil && *user.RefreshTokenExpireMinutes > 0 {
		refreshExpiresIn = int64(*user.RefreshTokenExpireMinutes * 60)
	}

	// 签发 refresh_token（绑定到该租户）
	refreshToken, err := strategy.GenerateRefreshToken(s.cfg, &strategy.RefreshTokenOptions{
		UserID:       claims.UserID,
		TenantID:     tenantID,
		UserPool:     strategy.UserPoolAdmin,
		TokenVersion: 0,
		ExpiresIn:    refreshExpiresIn,
	})
	if err != nil {
		return nil, err
	}

	// 记录 Token 到审计日志
	if s.tokenRecorder != nil {
		accessClaims, _ := strategy.ParseAccessToken(s.cfg, accessToken)
		refreshClaims, _ := strategy.ParseRefreshToken(s.cfg, refreshToken)
		accessJTI := ""
		refreshJTI := ""
		if accessClaims != nil {
			accessJTI = accessClaims.ID
		}
		if refreshClaims != nil {
			refreshJTI = refreshClaims.ID
		}
		_ = s.tokenRecorder.RecordSelectTenant(
			claims.UserID, user.Username, tenantID,
			accessToken, accessJTI, time.Now().Add(time.Duration(accessExpiresIn)*time.Second),
			refreshToken, refreshJTI, time.Now().Add(time.Duration(refreshExpiresIn)*time.Second),
			ip, userAgent,
		)
	}

	return &strategy.SelectTenantResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessExpiresIn,
	}, nil
}

// RefreshResponse 刷新响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Refresh 用 refresh_token 刷新 access_token + refresh_token
func (s *AuthService) Refresh(refreshTokenStr string, tenantID int64, ip, userAgent string) (*RefreshResponse, error) {
	// 解析 refresh_token
	refreshClaims, err := strategy.ParseRefreshToken(s.cfg, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	// 检查 refresh_token 是否在黑名单中（已被吊销或已刷新）
	if refreshClaims.ID != "" && s.blacklist.Contains(refreshClaims.ID) {
		return nil, errors.NewAuthError(errors.ErrTokenBlacklisted, "refresh_token 已失效")
	}

	// 检查用户状态
	userStatus, err := s.CheckUserStatus(refreshClaims.UserID)
	if err == nil && userStatus != 1 {
		return nil, errors.NewAuthError(errors.ErrAccountDisabled, "账号已禁用")
	}

	// 查询用户在该租户下的角色（刷新时也要包含角色，否则权限丢失）
	roleIDs := s.getUserRoles(refreshClaims.UserID, refreshClaims.TenantID)

	// 签发新的 access_token
	accessToken, err := strategy.GenerateAccessToken(s.cfg, &strategy.AccessTokenOptions{
		UserID:       refreshClaims.UserID,
		TenantID:     refreshClaims.TenantID,
		Roles:        roleIDs,
		UserPool:     refreshClaims.UserPool,
		TokenVersion: refreshClaims.TokenVersion,
	})
	if err != nil {
		return nil, err
	}

	// 签发新的 refresh_token（旧 refresh_token 通过黑名单失效）
	refreshToken, err := strategy.GenerateRefreshToken(s.cfg, &strategy.RefreshTokenOptions{
		UserID:       refreshClaims.UserID,
		TenantID:     refreshClaims.TenantID,
		UserPool:     refreshClaims.UserPool,
		TokenVersion: refreshClaims.TokenVersion,
	})
	if err != nil {
		return nil, err
	}

	// 旧 refresh_token 加入黑名单
	if refreshClaims.ID != "" {
		ttl := time.Until(refreshClaims.ExpiresAt.Time)
		if ttl <= 0 {
			ttl = time.Minute
		}
		_ = s.blacklist.Add(refreshClaims.ID, ttl)
		
		// 异步更新旧 token 记录状态为 USED（不阻塞刷新响应）
		if s.tokenRecorder != nil {
			oldRefreshJTI := refreshClaims.ID
			go func() {
				if err := s.tokenRecorder.MarkRefreshTokenUsed(oldRefreshJTI); err != nil {
					log.Printf("[auth] MarkRefreshTokenUsed failed: %v", err)
				}
			}()
		}
	}

	// 记录 Token 刷新
	if s.tokenRecorder != nil {
		newAccessClaims, _ := strategy.ParseAccessToken(s.cfg, accessToken)
		newRefreshClaims, _ := strategy.ParseRefreshToken(s.cfg, refreshToken)
		accessJTI := ""
		refreshJTI := ""
		if newAccessClaims != nil {
			accessJTI = newAccessClaims.ID
		}
		if newRefreshClaims != nil {
			refreshJTI = newRefreshClaims.ID
		}
		_ = s.tokenRecorder.RecordRefresh(
			refreshClaims.UserID, refreshClaims.TenantID,
			accessToken, accessJTI, time.Now().Add(s.cfg.JWT.AccessTokenTTL),
			refreshToken, refreshJTI, time.Now().Add(s.cfg.JWT.RefreshTokenTTL),
			ip, userAgent,
		)
	}

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTokenTTL.Seconds()),
	}, nil
}

// Logout 将 token 的 JTI 加入黑名单
func (s *AuthService) Logout(tokenStr string) error {
	log.Printf("[auth] Logout called")
	claims, err := strategy.ParseAccessToken(s.cfg, tokenStr)
	if err != nil {
		log.Printf("[auth] Logout: token parse error: %v", err)
		// 即使 token 过期也允许 logout
		return nil
	}

	log.Printf("[auth] Logout: jti=%s user_id=%d", claims.ID, claims.UserID)

	if claims.ID != "" {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl <= 0 {
			ttl = time.Minute // 已过期 token 保留 1 分钟防重放
		}
		_ = s.blacklist.Add(claims.ID, ttl)

		// 异步更新数据库中的 token 记录状态为 REVOKED（不阻塞登出响应）
		if s.tokenRecorder != nil {
			accessJTI := claims.ID
			go func() {
				if err := s.tokenRecorder.MarkAccessTokenRevoked(accessJTI); err != nil {
					log.Printf("[auth] MarkAccessTokenRevoked failed: %v", err)
				}
			}()
		}
	}
	return nil
}

// IsBlacklisted 检查 token JTI 是否在黑名单中
func (s *AuthService) IsBlacklisted(jti string) bool {
	if jti == "" {
		return false
	}
	return s.blacklist.Contains(jti)
}

// ParsePlatformToken 解析 platform_token（保留方法签名兼容）
func (s *AuthService) ParsePlatformToken(tokenStr string) (*strategy.PlatformClaims, error) {
	return strategy.ParsePlatformToken(s.cfg, tokenStr)
}

// ParseAccessToken 解析 access_token（保留方法签名兼容）
func (s *AuthService) ParseAccessToken(tokenStr string) (*strategy.AccessClaims, error) {
	return strategy.ParseAccessToken(s.cfg, tokenStr)
}

// getUserRoles 查询用户在指定租户下的有效角色 ID 列表
func (s *AuthService) getUserRoles(userID, tenantID int64) []int64 {
	var userRoles []model.UserRole
	now := time.Now()
	s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Find(&userRoles)

	roleIDs := make([]int64, 0, len(userRoles))
	for _, ur := range userRoles {
		if ur.EffectiveStart != nil && now.Before(*ur.EffectiveStart) {
			continue
		}
		if ur.EffectiveEnd != nil && now.After(*ur.EffectiveEnd) {
			continue
		}
		roleIDs = append(roleIDs, ur.RoleID)
	}
	return roleIDs
}
