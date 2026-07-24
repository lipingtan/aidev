package strategy

import (
	"errors"
	"net/http"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/model"
	"go-admin/app/user_auth/repository"
	"go-admin/app/user_auth/service"
	"go-admin/common/auth/config"
	authStrategy "go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SmsStrategy 短信验证码登录策略
// 实现 strategy.AuthenticationStrategy 接口
type SmsStrategy struct {
	smsService  *service.SmsService
	bizUserSvc  *service.BizUserService
	bizUserRepo repository.BizUserRepository
	cfg         *config.Config
	db          *gorm.DB
}

// NewSmsStrategy 创建 SmsStrategy 实例
func NewSmsStrategy(
	smsService *service.SmsService,
	bizUserSvc *service.BizUserService,
	bizUserRepo repository.BizUserRepository,
	cfg *config.Config,
	db *gorm.DB,
) *SmsStrategy {
	return &SmsStrategy{
		smsService:  smsService,
		bizUserSvc:  bizUserSvc,
		bizUserRepo: bizUserRepo,
		cfg:         cfg,
		db:          db,
	}
}

// GrantType 返回授权类型标识
func (s *SmsStrategy) GrantType() string { return "sms" }

// Authenticate 短信验证码认证
// 1. 解析请求体 {phone, code, tenant_code}
// 2. 验证 tenant_code → 查 admin_tenant → 获取 tenant_id，非 ACTIVE 则 400
// 3. 调用 smsService.VerifyCode(phone, tenant_code, code)
// 4. 按 (tenant_id, phone) 查 biz_user
//   - 存在 → 直接签发
//   - 不存在 → 创建 biz_user（唯一约束冲突时重试查询）→ 签发
//
// 5. 签发 access_token（UserPool=user, TokenVersion=user.TokenVersion）
// 6. 将签发结果写入 c.Set("login_response", ...)
func (s *SmsStrategy) Authenticate(c *gin.Context) (*authStrategy.AuthResult, error) {
	var req dto.SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "请求参数错误: " + err.Error()})
		return nil, err
	}

	// 验证 tenant_code → 查 admin_tenant
	tenantID, err := s.validateTenant(req.TenantCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return nil, err
	}

	// 校验验证码
	if err := s.smsService.VerifyCode(req.Phone, req.TenantCode, req.Code); err != nil {
		if errors.Is(err, service.ErrCodeInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "验证码错误"})
			return nil, err
		}
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "验证码已过期或不存在"})
		return nil, err
	}

	// 查找或创建用户
	user, err := s.findOrCreateUser(tenantID, req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "用户处理失败"})
		return nil, err
	}

	// 签发 access_token
	tokenTTL := s.cfg.JWT.AccessTokenTTL
	accessToken, err := authStrategy.GenerateAccessToken(s.cfg, &authStrategy.AccessTokenOptions{
		UserID:       user.ID,
		TenantID:     tenantID,
		Roles:        nil,
		UserPool:     authStrategy.UserPoolUser,
		TokenVersion: user.TokenVersion,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "签发令牌失败"})
		return nil, err
	}

	// 将登录响应写入上下文
	c.Set("login_response", &dto.SmsLoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(tokenTTL.Seconds()),
	})

	return &authStrategy.AuthResult{
		UserID:       user.ID,
		TenantID:     tenantID,
		UserPool:     authStrategy.UserPoolUser,
		Roles:        nil,
		TokenVersion: user.TokenVersion,
	}, nil
}

// validateTenant 验证 tenant_code 有效性
// 查询 admin_tenant 表，租户不存在或非 ACTIVE(status=1) 返回错误
func (s *SmsStrategy) validateTenant(tenantCode string) (int64, error) {
	var result struct {
		ID     int64
		Status int
	}
	err := s.db.Table("admin_tenant").
		Select("id, status").
		Where("tenant_code = ?", tenantCode).
		Scan(&result).Error
	if err != nil {
		return 0, errors.New("租户查询失败")
	}
	if result.ID == 0 {
		return 0, errors.New("无效的租户编码")
	}
	if result.Status != 1 {
		return 0, errors.New("租户已禁用或不可用")
	}
	return result.ID, nil
}

// findOrCreateUser 按 (tenant_id, phone) 查找用户，不存在则自动创建
// 唯一约束冲突时重试查询（upsert 语义）
func (s *SmsStrategy) findOrCreateUser(tenantID int64, phone string) (*model.BizUser, error) {
	// 先尝试查找
	user, err := s.bizUserRepo.FindByTenantPhone(s.db, tenantID, phone)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 不存在，创建新用户
	newUser, err := s.bizUserSvc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: tenantID,
		Phone:    phone,
		Nickname: "用户" + phone[len(phone)-4:],
	})
	if err != nil {
		// 唯一约束冲突 → 重新查询
		if errors.Is(err, service.ErrDuplicatePhone) {
			user, retryErr := s.bizUserRepo.FindByTenantPhone(s.db, tenantID, phone)
			if retryErr != nil {
				return nil, retryErr
			}
			return user, nil
		}
		return nil, err
	}
	return newUser, nil
}
