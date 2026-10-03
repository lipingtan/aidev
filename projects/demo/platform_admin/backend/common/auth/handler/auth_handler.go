package handler

import (
	"fmt"
	"log"
	"strings"

	"go-admin/common/auth/errors"
	"go-admin/common/auth/middleware"
	"go-admin/common/auth/service"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
)

// AuthHandler 认证 HTTP Handler
type AuthHandler struct {
	authSvc        *service.AuthService
	userSvc        *service.UserService
	router         *strategy.StrategyRouter
	loginLogSvc    *service.LoginLogService
	tokenRecorder  *service.TokenRecordService
	loginGuard     *service.LoginAttemptGuard // v2-cr11: 登录失败 IP 黑名单防护
	adminConfigSvc *service.AdminConfigService // v2-cr11: 阈值/时长配置读取
}

// NewAuthHandler 构造认证 Handler
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	h := &AuthHandler{
		authSvc: authSvc,
		router:  strategy.NewStrategyRouter(),
	}
	// 注册密码策略
	h.router.Register(strategy.NewPasswordStrategy(func(username, password, ip, userAgent string, rememberMe bool) (*strategy.LoginResponse, error) {
		return authSvc.LoginWithRememberMe(username, password, ip, userAgent, rememberMe)
	}))
	// 注册 OAuth2/LDAP 骨架策略（CR-8）
	h.router.Register(&strategy.OAuth2Strategy{})
	h.router.Register(&strategy.LDAPStrategy{})
	return h
}

// SetTokenRecordService 注入令牌操作审计日志服务
func (h *AuthHandler) SetTokenRecordService(svc *service.TokenRecordService) {
	h.tokenRecorder = svc
}

// SetLoginLogService 注入登录日志服务（用于写入登录记录）
func (h *AuthHandler) SetLoginLogService(svc *service.LoginLogService) {
	h.loginLogSvc = svc
}

// SetLoginGuard 注入登录防护守卫（v2-cr11：失败计数 + IP 黑名单）
func (h *AuthHandler) SetLoginGuard(guard *service.LoginAttemptGuard) {
	h.loginGuard = guard
}

// SetAdminConfigService 注入系统配置服务（v2-cr11：阈值/时长配置读取）
func (h *AuthHandler) SetAdminConfigService(svc *service.AdminConfigService) {
	h.adminConfigSvc = svc
}

// 登录失败封禁配置项 key（系统配置 admin_config，平台级）
const (
	configKeyLoginFailBanThreshold = "auth.login_fail_ban_threshold"      // 一级阈值（次），0=关闭
	configKeyLoginFailIPBanThreshold = "auth.login_fail_ip_ban_threshold" // 二级阈值（同 IP 组合封禁数），0=关闭
	configKeyLoginFailBanHours     = "auth.login_fail_ban_hours"          // 封禁时长（小时）
)

// 登录失败计数的默认值（admin_config 无记录时生效）
const (
	defaultLoginFailBanThreshold   = 10
	defaultLoginFailIPBanThreshold = 5
	defaultLoginFailBanHours       = 24
)

// SetUserService 注入用户服务（用于修改密码）
func (h *AuthHandler) SetUserService(svc *service.UserService) {
	h.userSvc = svc
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
	RememberMe  bool   `json:"remember_me"` // 记住登录：refresh_token 延长至 30 天（默认 7 天；用户级 TTL 配置仍优先）
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

// handlePasswordLogin 处理密码登录
// v2-cr11: 黑名单校验（验证码后、密码校验前）→ 失败计数 → 达标封禁
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

	clientIP := c.ClientIP()

	// v2-cr11: 两级黑名单前置校验（密码 bcrypt 计算之前，防刷省算力）：
	// 一级=整 IP 封禁（该 IP 任意用户名全拒）；二级=IP+用户名 组合封禁（仅该组合拒）
	if h.loginGuard != nil {
		if banned, remaining := h.loginGuard.CheckBanned(clientIP, req.Username); banned {
			remainingMin := int(remaining.Minutes()) + 1
			message := fmt.Sprintf("登录失败次数过多，已被封禁，剩余约 %d 分钟，请稍后重试或联系管理员", remainingMin)
			// 拦截也写登录日志（可审计）
			if h.loginLogSvc != nil {
				h.loginLogSvc.CreateAsync(0, req.Username, clientIP, c.Request.UserAgent(), 0, "已被封禁，登录拦截")
			}
			Error(c, errors.NewAuthError(errors.ErrIPBlacklisted, message))
			return
		}
	}

	resp, err := h.authSvc.LoginWithRememberMe(req.Username, req.Password, clientIP, c.Request.UserAgent(), req.RememberMe)

	// 写登录日志
	if h.loginLogSvc != nil {
		status := 1
		message := "登录成功"
		if err != nil {
			status = 0
			message = "用户名或密码错误"
		}
		userID := h.authSvc.GetUserIDByUsername(req.Username)
		h.loginLogSvc.CreateAsync(userID, req.Username, clientIP, c.Request.UserAgent(), status, message)
	}

	if err != nil {
		// v2-cr11: 凭证错误时累计失败计数，达标触发两级封禁
		h.recordLoginFailure(clientIP, req.Username)
		Error(c, err)
		return
	}

	// v2-cr11: 登录成功清零失败计数
	if h.loginGuard != nil {
		h.loginGuard.Reset(clientIP, req.Username)
	}

	Success(c, resp)
}

// recordLoginFailure 记录一次登录失败；组合达标触发一级封禁（IP+用户名），
// 同 IP 组合封禁数达到二级阈值时升级封禁整个 IP。阈值=0 关闭对应层级。
// 封禁失败仅记日志（降级放行，不影响错误响应）。
func (h *AuthHandler) recordLoginFailure(ip, username string) {
	if h.loginGuard == nil {
		return
	}
	threshold := defaultLoginFailBanThreshold
	if h.adminConfigSvc != nil {
		threshold = h.adminConfigSvc.ResolveInt(0, configKeyLoginFailBanThreshold, defaultLoginFailBanThreshold)
	}
	if threshold <= 0 {
		return // 一级关闭（RG-5）
	}

	count := h.loginGuard.RecordFailure(ip, username)
	if count < threshold {
		return
	}

	hours := defaultLoginFailBanHours
	if h.adminConfigSvc != nil {
		hours = h.adminConfigSvc.ResolveInt(0, configKeyLoginFailBanHours, defaultLoginFailBanHours)
	}

	// 一级：封 IP+用户名 组合（幂等）
	h.loginGuard.BanCombo(ip, username, count, hours)

	// 二级：同 IP 生效中的组合封禁数达到阈值 → 升级封整个 IP
	ipThreshold := defaultLoginFailIPBanThreshold
	if h.adminConfigSvc != nil {
		ipThreshold = h.adminConfigSvc.ResolveInt(0, configKeyLoginFailIPBanThreshold, defaultLoginFailIPBanThreshold)
	}
	if ipThreshold > 0 && h.loginGuard.CountActiveComboBans(ip) >= ipThreshold {
		h.loginGuard.BanIP(ip, count, hours)
	}
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

	resp, err := h.authSvc.SelectTenant(tokenStr, req.TenantID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// RefreshRequest 刷新请求
type RefreshRequest struct {
	TenantID     int64  `json:"tenant_id,string"`
	RefreshToken string `json:"refresh_token"`
}

// Refresh 刷新 access_token + refresh_token
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "参数错误: " + err.Error()})
		return
	}

	refreshToken := req.RefreshToken
	if refreshToken == "" {
		// 兼容旧版：从 Authorization header 获取
		refreshToken = extractBearerToken(c)
	}
	if refreshToken == "" {
		Error(c, &errUnauthorized{message: "缺少 refresh_token"})
		return
	}

	resp, err := h.authSvc.Refresh(refreshToken, req.TenantID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	log.Printf("[auth-handler] Logout called from %s", c.ClientIP())
	tokenStr := extractBearerToken(c)
	if tokenStr == "" {
		log.Printf("[auth-handler] Logout: no token found")
		Error(c, &errUnauthorized{message: "缺少认证信息"})
		return
	}

	log.Printf("[auth-handler] Logout: calling authSvc.Logout")
	if err := h.authSvc.Logout(tokenStr); err != nil {
		log.Printf("[auth-handler] Logout: authSvc.Logout error: %v", err)
		Error(c, err)
		return
	}

	// 记录令牌操作审计日志
	if h.tokenRecorder != nil {
		log.Printf("[auth-handler] Logout: recording logout audit")
		_ = h.tokenRecorder.RecordLogout(0, 0, "access", "", c.ClientIP(), c.Request.UserAgent())
	}

	log.Printf("[auth-handler] Logout: success")
	Success(c, nil)
}

// ChangeOwnPassword 用户修改自己的密码
// PUT /api/v1/common/password
func (h *AuthHandler) ChangeOwnPassword(c *gin.Context) {
	// 从认证上下文获取当前用户 ID
	authCtx := middleware.GetAuthContext(c)
	if authCtx == nil || authCtx.UserID == 0 {
		Error(c, &errUnauthorized{message: "缺少用户信息"})
		return
	}

	var req service.ChangeOwnPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, &errBadRequest{message: "请求参数无效: " + err.Error()})
		return
	}

	if err := h.userSvc.ChangeOwnPassword(authCtx.UserID, &req); err != nil {
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

// errNotFound 内部错误类型，用于资源不存在
type errNotFound struct {
	message string
}

func (e *errNotFound) Error() string {
	return e.message
}
