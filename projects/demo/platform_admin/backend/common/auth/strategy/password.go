package strategy

import (
	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
)

// LoginFunc 密码登录执行函数（由外部注入，避免循环依赖）
type LoginFunc func(username, password string) (*LoginResponse, error)

// PasswordStrategy 管理端密码登录策略
type PasswordStrategy struct {
	loginFn LoginFunc
}

// NewPasswordStrategy 创建密码登录策略
func NewPasswordStrategy(loginFn LoginFunc) *PasswordStrategy {
	return &PasswordStrategy{loginFn: loginFn}
}

// GrantType 返回策略类型标识
func (s *PasswordStrategy) GrantType() string {
	return "password"
}

// passwordLoginRequest 密码登录请求体
type passwordLoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaKey  string `json:"captcha_key"`
	CaptchaCode string `json:"captcha_code"`
}

// Authenticate 执行密码认证
func (s *PasswordStrategy) Authenticate(c *gin.Context) (*AuthResult, error) {
	var req passwordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, &PasswordStrategyError{Message: "参数错误: " + err.Error(), IsBadRequest: true}
	}

	// 验证码校验（非空时必须通过）
	if req.CaptchaKey != "" || req.CaptchaCode != "" {
		if req.CaptchaKey == "" || req.CaptchaCode == "" {
			return nil, &PasswordStrategyError{Message: "请输入验证码", IsBadRequest: true}
		}
		if !captcha.Verify(req.CaptchaKey, req.CaptchaCode, true) {
			return nil, &PasswordStrategyError{Message: "验证码错误", IsBadRequest: true}
		}
	}

	// 调用注入的登录函数
	resp, err := s.loginFn(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	// 将登录响应暂存到 gin.Context 供 handler 层使用
	c.Set("login_response", resp)

	return &AuthResult{
		UserPool: UserPoolAdmin,
	}, nil
}

// PasswordStrategyError 密码策略错误
type PasswordStrategyError struct {
	Message      string
	IsBadRequest bool
}

func (e *PasswordStrategyError) Error() string {
	return e.Message
}
