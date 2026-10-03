package handler

import (
	"errors"
	"net/http"
	"strings"

	"go-admin/app/user_auth/dto"
	"go-admin/app/user_auth/model"
	"go-admin/app/user_auth/service"
	"go-admin/common/auth/config"
	"go-admin/common/auth/middleware"
	authStrategy "go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserAuthHandler C端用户认证 HTTP Handler
type UserAuthHandler struct {
	smsService  *service.SmsService
	bizUserSvc  *service.BizUserService
	menuSvc     *service.UserMenuService
	cfg         *config.Config
	db          *gorm.DB
}

// NewUserAuthHandler 创建 UserAuthHandler 实例
func NewUserAuthHandler(
	smsService *service.SmsService,
	bizUserSvc *service.BizUserService,
	menuSvc *service.UserMenuService,
	cfg *config.Config,
	db *gorm.DB,
) *UserAuthHandler {
	return &UserAuthHandler{
		smsService: smsService,
		bizUserSvc: bizUserSvc,
		menuSvc:    menuSvc,
		cfg:        cfg,
		db:         db,
	}
}

// SendCode POST /api/v1/user/public/send-code
// 请求: {"phone": "...", "tenant_code": "..."} （tenant_code 可选，为空时从 Host 识别）
// 响应: {"code": 200, "data": {"expires_in": 300}}
// 限频: 429 + {"retry_after": N}
func (h *UserAuthHandler) SendCode(c *gin.Context) {
	var req dto.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 解析租户：优先使用请求体中的 tenant_code，为空时从 Host 识别
	tenantCode, err := h.resolveTenantCode(c, req.TenantCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	// 发送验证码
	_, err = h.smsService.SendCode(req.Phone, tenantCode)
	if err != nil {
		// 限频错误
		var tooFreqErr *service.TooFrequentError
		if errors.As(err, &tooFreqErr) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":        42900,
				"data":        nil,
				"message":     "发送过于频繁",
				"retry_after": tooFreqErr.RemainSeconds,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "发送验证码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    dto.SendCodeResponse{ExpiresIn: 300},
		"message": "success",
	})
}

// Login POST /api/v1/user/public/login
// 请求: {"account": "手机号或邮箱", "code": "...", "tenant_code": "..."} （tenant_code 可选，为空时从 Host 识别）
// 成功: {"code": 200, "data": {"access_token": "...", "expires_in": N}}
// 验证码错误: 401
// 租户识别失败: 400
func (h *UserAuthHandler) Login(c *gin.Context) {
	var req dto.SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 解析租户：优先使用请求体中的 tenant_code，为空时从 Host 识别
	tenantCode, err := h.resolveTenantCode(c, req.TenantCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	// 验证 tenant_code → 查 admin_tenant 获取 tenant_id
	tenantID, err := h.resolveTenant(tenantCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "data": nil, "message": err.Error()})
		return
	}

	// 校验验证码（account 可以是手机号或邮箱）
	if err := h.smsService.VerifyCode(req.Account, tenantCode, req.Code); err != nil {
		if errors.Is(err, service.ErrCodeInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "验证码错误"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "验证码已过期或不存在"})
		return
	}

	// 查找或创建用户（支持手机号或邮箱）
	user, err := h.findOrCreateUserByAccount(tenantID, req.Account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "用户处理失败"})
		return
	}

	// 签发 access_token
	tokenTTL := h.cfg.JWT.AccessTokenTTL
	accessToken, err := authStrategy.GenerateAccessToken(h.cfg, &authStrategy.AccessTokenOptions{
		UserID:       user.ID,
		TenantID:     tenantID,
		Roles:        nil,
		UserPool:     authStrategy.UserPoolUser,
		TokenVersion: user.TokenVersion,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "签发令牌失败"})
		return
	}

	// 签发 refresh_token
	refreshToken, err := authStrategy.GenerateRefreshToken(h.cfg, &authStrategy.RefreshTokenOptions{
		UserID:       user.ID,
		TenantID:     tenantID,
		UserPool:     authStrategy.UserPoolUser,
		TokenVersion: user.TokenVersion,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "签发刷新令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": dto.SmsLoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(tokenTTL.Seconds()),
		},
		"message": "success",
	})
}

// Logout POST /api/v1/user/auth/logout
// 需要 JWT(user) 认证
// 调用 bizUserService.ForceLogout(userID)
func (h *UserAuthHandler) Logout(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	if err := h.bizUserSvc.ForceLogout(authCtx.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "登出失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil, "message": "success"})
}

// GetMenu GET /api/v1/user/menu
// 从 AuthContext 获取 tenantID，返回该租户 platform=user 的菜单树
func (h *UserAuthHandler) GetMenu(c *gin.Context) {
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "data": nil, "message": "未认证"})
		return
	}

	tree, err := h.menuSvc.GetUserMenu(authCtx.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "data": nil, "message": "获取菜单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tree, "message": "success"})
}

// resolveTenantCode 解析租户编码
// 1. 如果请求体中 tenant_code 非空，直接使用并验证
// 2. 如果 tenant_code 为空，从 Host 头识别租户
func (h *UserAuthHandler) resolveTenantCode(c *gin.Context, tenantCodeFromBody string) (string, error) {
	// 优先使用请求体中的 tenant_code
	if tenantCodeFromBody != "" {
		// 验证 tenant_code 有效性
		if err := h.validateTenantCode(tenantCodeFromBody); err != nil {
			return "", err
		}
		return tenantCodeFromBody, nil
	}

	// 从 Host 头获取域名，查 tenant_domain 表
	host := c.Request.Host
	// 去掉端口号（如果有）
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// 查 tenant_domain 表获取 tenant_code
	var result struct {
		TenantCode string
	}
	err := h.db.Table("tenant_domain").
		Select("t.tenant_code").
		Joins("JOIN admin_tenant t ON t.id = tenant_domain.tenant_id").
		Where("tenant_domain.domain = ? AND tenant_domain.deleted_at IS NULL AND t.status = 1", host).
		Scan(&result).Error
	if err != nil {
		return "", errors.New("域名查询失败")
	}
	if result.TenantCode == "" {
		return "", errors.New("无法识别租户，请确认请求域名或传入 tenant_code")
	}

	return result.TenantCode, nil
}

// validateTenantCode 验证 tenant_code 是否有效且为 ACTIVE 状态
func (h *UserAuthHandler) validateTenantCode(tenantCode string) error {
	var result struct {
		ID     int64
		Status int
	}
	err := h.db.Table("admin_tenant").
		Select("id, status").
		Where("tenant_code = ?", tenantCode).
		Scan(&result).Error
	if err != nil {
		return errors.New("租户查询失败")
	}
	if result.ID == 0 {
		return errors.New("无效的租户编码")
	}
	if result.Status != 1 {
		return errors.New("租户已禁用或不可用")
	}
	return nil
}

// resolveTenant 解析 tenant_code 返回 tenant_id
func (h *UserAuthHandler) resolveTenant(tenantCode string) (int64, error) {
	var result struct {
		ID     int64
		Status int
	}
	err := h.db.Table("admin_tenant").
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
func (h *UserAuthHandler) findOrCreateUser(tenantID int64, phone string) (*model.BizUser, error) {
	user, err := h.bizUserSvc.GetBizUserByTenantPhone(tenantID, phone)
	if err == nil {
		return user, nil
	}

	// 不存在，创建新用户
	nickname := "用户" + phone[len(phone)-4:]
	newUser, err := h.bizUserSvc.CreateBizUser(&dto.CreateBizUserRequest{
		TenantID: tenantID,
		Phone:    phone,
		Nickname: nickname,
	})
	if err != nil {
		// 唯一约束冲突 → 重新查询
		if errors.Is(err, service.ErrDuplicatePhone) {
			return h.bizUserSvc.GetBizUserByTenantPhone(tenantID, phone)
		}
		return nil, err
	}
	return newUser, nil
}

// findOrCreateUserByAccount 按账号（手机号或邮箱）查找用户，不存在则自动创建
func (h *UserAuthHandler) findOrCreateUserByAccount(tenantID int64, account string) (*model.BizUser, error) {
	// 先尝试查找已有用户
	user, err := h.bizUserSvc.GetBizUserByTenantAccount(tenantID, account)
	if err == nil {
		return user, nil
	}

	// 不存在，创建新用户
	isEmail := isEmailFormat(account)
	
	var phone, email string
	var nickname string
	
	if isEmail {
		email = account
		// 邮箱用户，昵称取 @ 前面部分
		if idx := strings.Index(account, "@"); idx > 0 {
			nickname = "用户" + account[:idx]
			if len(nickname) > 20 {
				nickname = nickname[:20]
			}
		} else {
			nickname = "用户"
		}
	} else {
		phone = account
		// 手机用户，昵称取后4位
		if len(account) >= 4 {
			nickname = "用户" + account[len(account)-4:]
		} else {
			nickname = "用户" + account
		}
	}
	
	newUser, err := h.bizUserSvc.CreateBizUserByAccount(&dto.CreateBizUserByAccountRequest{
		TenantID: tenantID,
		Phone:    phone,
		Email:    email,
		Nickname: nickname,
	})
	if err != nil {
		// 唯一约束冲突 → 重新查询
		if errors.Is(err, service.ErrDuplicatePhone) || errors.Is(err, service.ErrDuplicateEmail) {
			return h.bizUserSvc.GetBizUserByTenantAccount(tenantID, account)
		}
		return nil, err
	}
	return newUser, nil
}

// isEmailFormat 判断是否为邮箱格式
func isEmailFormat(s string) bool {
	return len(s) > 3 && strings.Contains(s, "@") && strings.Contains(s, ".")
}
