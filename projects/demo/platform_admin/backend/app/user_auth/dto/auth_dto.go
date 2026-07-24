package dto

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone      string `json:"phone" binding:"required"`
	TenantCode string `json:"tenant_code" binding:"required"`
}

// SendCodeResponse 发送验证码响应
type SendCodeResponse struct {
	ExpiresIn int `json:"expires_in"`
}

// SmsLoginRequest 短信验证码登录请求
type SmsLoginRequest struct {
	Phone      string `json:"phone" binding:"required"`
	Code       string `json:"code" binding:"required"`
	TenantCode string `json:"tenant_code" binding:"required"`
}

// SmsLoginResponse 短信验证码登录响应
type SmsLoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
