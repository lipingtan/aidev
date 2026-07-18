package handler

import (
	"strings"

	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
)

// AuthHandler 认证 HTTP Handler
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler 构造认证 Handler
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
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
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CaptchaKey  string `json:"captcha_key"`
	CaptchaCode string `json:"captcha_code"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
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
