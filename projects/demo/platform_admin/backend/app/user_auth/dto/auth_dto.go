package dto

// SendCodeRequest 发送验证码请求
// C 端请求时 tenant_code 可选，后端从 Host 自动识别租户
// 管理端或显式指定时可传 tenant_code
type SendCodeRequest struct {
	Phone      string `json:"phone" binding:"required"`
	TenantCode string `json:"tenant_code"` // 可选，为空时从 Host 识别
}

// SendCodeResponse 发送验证码响应
type SendCodeResponse struct {
	ExpiresIn int `json:"expires_in"`
}

// SmsLoginRequest 短信验证码登录请求（支持手机号或邮箱）
// C 端请求时 tenant_code 可选，后端从 Host 自动识别租户
// 管理端或显式指定时可传 tenant_code
type SmsLoginRequest struct {
	Account    string `json:"account" binding:"required"` // 手机号或邮箱
	Code       string `json:"code" binding:"required"`
	TenantCode string `json:"tenant_code"` // 可选，为空时从 Host 识别
}

// SmsLoginResponse 短信验证码登录响应
type SmsLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
