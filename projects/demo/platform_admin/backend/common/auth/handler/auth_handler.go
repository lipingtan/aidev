package handler

import (
	"strings"

	"go-admin/common/auth/service"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
)

// AuthHandler 认证 HTTP Handler
type AuthHandler struct {
	authSvc *service.AuthService
	router  *strategy.StrategyRouter
}

// NewAuthHandler 构造认证 Handler
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	h := &AuthHandler{
		authSvc: authSvc,
		router:  strategy.NewStrategyRouter(),
	}
	// 注册密码策略
	h.router.Register(strategy.NewPasswordStrategy(authSvc.Login))
	return h
}

// GetStrategyRouter 返回策略路由器（供外部注册新策略）
func (h *AuthHandler) GetStrategyRouter() *strategy.StrategyRouter {
	return h.router
}

// RegisterRoutes 注册 /auth 路由组
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.GET("/captcha", h.Captcha)
		auth.POST("/login", h.Login)
		auth.POST("/tenant/select", h.SelectTenant)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
	}
}

// Captcha 获取图形验证码
func (h *AuthHandler) Captcha(c *gin.Context) {
	id, b64s, _, err := captcha.DriverDigitFunc()
	if err != nil {
		Error(c, &errBadRequest{message: "验证码生成失败"})
		return
	}
	Success(c, gin.H{
		"captcha_key":   id,
		"captcha_image": b64s,
	})
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaKey  string `json:"captcha_key"`
	CaptchaCode string `json:"captcha_code"`
	GrantType   string `json:"grant_type"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	// grant_type 为空或 "password" → 执行现有密码登录逻辑
	if req.GrantType == "" || req.GrantType == "password" {
		h.handlePasswordLogin(c, &req)
		return
	}

	// 其他 grant_type → 通过策略路由分发
	s, err := h.router.Route(req.GrantType)
	if err != nil {
		Error(c, &errBadRequest{message: "不支持的认证类型: " + req.GrantType})
		return
	}

	_, authErr := s.Authenticate(c)
	if authErr != nil {
		Error(c, authErr)
		return
	}

	// 从 context 获取登录响应（由策略写入）
	respVal, exists := c.Get("login_response")
	if !exists {
		Error(c, &errBadRequest{message: "认证策略未返回结果"})
		return
	}
	Success(c, respVal)
}

// handlePasswordLogin 处理密码登录（保持原有逻辑不变）
func (h *AuthHandler) handlePasswordLogin(c *gin.Context, req *LoginRequest) {
	if req.Username == "" || req.Password == "" {
		Error(c, &errBadRequest{message: "用户名和密码不能为空"})
		return
	}

	// 验证码校验（非空时必须通过）
	if req.CaptchaKey != "" || req.CaptchaCode != "" {
		if req.CaptchaKey == "" || req.CaptchaCode == "" {
			Error(c, &errBadRequest{message: "请输入验证码"})
			return
		}
		if !captcha.Verify(req.CaptchaKey, req.CaptchaCode, true) {
			Error(c, &errBadRequest{message: "验证码错误"})
			return
		}
	}

	resp, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// SelectTenantRequest 选择租户请求
type SelectTenantRequest struct {
	TenantID int64 `json:"tenant_id,string" binding:"required"`
}

// SelectTenant 选择/切换租户
func (h *AuthHandler) SelectTenant(c *gin.Context) {
	var req SelectTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	tokenStr := extractBearerToken(c)
	if tokenStr == "" {
		Error(c, &errUnauthorized{message: "缺少认证信息"})
		return
	}

	resp, err := h.authSvc.SelectTenant(tokenStr, req.TenantID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// RefreshRequest 刷新请求
type RefreshRequest struct {
	TenantID int64 `json:"tenant_id,string" binding:"required"`
}

// Refresh 刷新 access_token
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	tokenStr := extractBearerToken(c)
	if tokenStr == "" {
		Error(c, &errUnauthorized{message: "缺少认证信息"})
		return
	}

	resp, err := h.authSvc.Refresh(tokenStr, req.TenantID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	tokenStr := extractBearerToken(c)
	if tokenStr == "" {
		Error(c, &errUnauthorized{message: "缺少认证信息"})
		return
	}

	if err := h.authSvc.Logout(tokenStr); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// extractBearerToken 从 Authorization header 提取 Bearer token
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// errBadRequest 内部错误类型，用于参数校验错误
type errBadRequest struct {
	message string
}

func (e *errBadRequest) Error() string {
	return e.message
}

// errUnauthorized 内部错误类型，用于未认证错误
type errUnauthorized struct {
	message string
}

func (e *errUnauthorized) Error() string {
	return e.message
}
