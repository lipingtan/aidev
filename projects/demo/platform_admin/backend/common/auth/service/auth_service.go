package service

import (
	"sync"
	"time"

	"go-admin/common/auth/config"
	"go-admin/common/auth/errors"
	"go-admin/common/auth/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TenantInfo 租户简要信息（嵌入 JWT claims）
type TenantInfo struct {
	ID   int64  `json:"id,string"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// PlatformClaims platform_token 的 JWT claims
type PlatformClaims struct {
	jwt.RegisteredClaims
	UserID  int64        `json:"user_id"`
	Tenants []TenantInfo `json:"tenants"`
}

// AccessClaims access_token 的 JWT claims
type AccessClaims struct {
	jwt.RegisteredClaims
	UserID   int64   `json:"user_id"`
	TenantID int64   `json:"tenant_id"`
	Roles    []int64 `json:"roles"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	TokenType     string       `json:"token_type"`               // "platform" 或 "access"
	Token         string       `json:"token"`                    // platform_token 或 access_token
	AccessToken   string       `json:"access_token,omitempty"`   // 单租户时直接返回
	Tenants       []TenantInfo `json:"tenants,omitempty"`        // 可用租户列表
	ExpiresIn     int64        `json:"expires_in"`               // 过期秒数
	PlatformToken string       `json:"platform_token,omitempty"` // 单租户时也返回 platform_token 用于后续刷新
}

// SelectTenantResponse 选择租户响应
type SelectTenantResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// AuthService 认证服务
type AuthService struct {
	db        *gorm.DB
	cfg       *config.Config
	blacklist sync.Map // token JTI -> 过期时间
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
	err := s.db.Table("admin_user").Where("id = ?", userID).Pluck("status", &status).Error
	return status, err
}

// NewAuthService 构造认证服务
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
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

	tenantInfos := make([]TenantInfo, len(tenants))
	for i, t := range tenants {
		tenantInfos[i] = TenantInfo{
			ID:   t.ID,
			Code: t.TenantCode,
			Name: t.Name,
		}
	}

	// 单租户优化：直接签发 access_token
	if len(tenantInfos) == 1 {
		platformToken, err := s.generatePlatformToken(user.ID, tenantInfos)
		if err != nil {
			return nil, err
		}

		accessToken, err := s.generateAccessToken(user.ID, tenantInfos[0].ID)
		if err != nil {
			return nil, err
		}

		return &LoginResponse{
			TokenType:     "access",
			Token:         accessToken,
			AccessToken:   accessToken,
			PlatformToken: platformToken,
			Tenants:       tenantInfos,
			ExpiresIn:     int64(s.cfg.JWT.AccessTokenTTL.Seconds()),
		}, nil
	}

	// 多租户：签发 platform_token
	platformToken, err := s.generatePlatformToken(user.ID, tenantInfos)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		TokenType: "platform",
		Token:     platformToken,
		Tenants:   tenantInfos,
		ExpiresIn: int64(s.cfg.JWT.PlatformTokenTTL.Seconds()),
	}, nil
}

// SelectTenant 选择租户，签发 access_token
func (s *AuthService) SelectTenant(platformTokenStr string, tenantID int64) (*SelectTenantResponse, error) {
	// 解析 platform_token
	claims, err := s.ParsePlatformToken(platformTokenStr)
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

	// 签发 access_token
	accessToken, err := s.generateAccessToken(claims.UserID, tenantID)
	if err != nil {
		return nil, err
	}

	return &SelectTenantResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.cfg.JWT.AccessTokenTTL.Seconds()),
	}, nil
}

// Refresh 用 platform_token 刷新 access_token
func (s *AuthService) Refresh(platformTokenStr string, tenantID int64) (*SelectTenantResponse, error) {
	// 逻辑与 SelectTenant 相同
	return s.SelectTenant(platformTokenStr, tenantID)
}

// Logout 将 token 加入黑名单
func (s *AuthService) Logout(tokenStr string) error {
	// 解析 access_token 获取 JTI 和过期时间
	claims, err := s.ParseAccessToken(tokenStr)
	if err != nil {
		// 即使 token 过期也允许 logout
		return nil
	}

	if claims.ID != "" {
		s.blacklist.Store(claims.ID, claims.ExpiresAt.Time)
	}

	return nil
}

// IsBlacklisted 检查 token 是否在黑名单中
func (s *AuthService) IsBlacklisted(jti string) bool {
	if jti == "" {
		return false
	}
	val, ok := s.blacklist.Load(jti)
	if !ok {
		return false
	}
	// 如果已过期，从黑名单中移除
	expTime := val.(time.Time)
	if time.Now().After(expTime) {
		s.blacklist.Delete(jti)
		return false
	}
	return true
}

// ParsePlatformToken 解析 platform_token
func (s *AuthService) ParsePlatformToken(tokenStr string) (*PlatformClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &PlatformClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "platform_token 无效或已过期")
	}

	claims, ok := token.Claims.(*PlatformClaims)
	if !ok || !token.Valid {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "platform_token 无效")
	}

	return claims, nil
}

// ParseAccessToken 解析 access_token
func (s *AuthService) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "access_token 无效或已过期")
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.NewAuthError(errors.ErrTokenExpired, "access_token 无效")
	}

	return claims, nil
}

// generatePlatformToken 生成 platform_token
func (s *AuthService) generatePlatformToken(userID int64, tenants []TenantInfo) (string, error) {
	now := time.Now()
	claims := &PlatformClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    s.cfg.JWT.Issuer,
			Subject:   "platform",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWT.PlatformTokenTTL)),
		},
		UserID:  userID,
		Tenants: tenants,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// generateAccessToken 生成 access_token
func (s *AuthService) generateAccessToken(userID, tenantID int64) (string, error) {
	// 查询用户在该租户下的角色
	var userRoles []model.UserRole
	now := time.Now()
	s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Find(&userRoles)

	// 过滤有效时间窗口内的角色
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

	claims := &AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    s.cfg.JWT.Issuer,
			Subject:   "access",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWT.AccessTokenTTL)),
		},
		UserID:   userID,
		TenantID: tenantID,
		Roles:    roleIDs,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}
