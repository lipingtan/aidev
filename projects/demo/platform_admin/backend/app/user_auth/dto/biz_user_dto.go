package dto

// CreateBizUserRequest 创建C端用户请求
// TenantID 由 handler 从 AuthContext 自动注入，前端无需传
// Phone 和 Email 至少填一个
type CreateBizUserRequest struct {
	TenantID int64  `json:"tenant_id,string"` // handler 层从 AuthContext 注入，binding 不强制
	Phone    string `json:"phone"`            // 手机号（与 Email 至少填一个）
	Email    string `json:"email"`            // 邮箱（与 Phone 至少填一个）
	Password string `json:"password"`         // 可选
	Nickname string `json:"nickname"`
	// Token 过期时间（分钟），NULL 时使用全局配置
	AccessTokenExpireMinutes  *int `json:"access_token_expire_minutes"`
	RefreshTokenExpireMinutes *int `json:"refresh_token_expire_minutes"`
}

// CreateBizUserByAccountRequest 按账号（手机号或邮箱）创建C端用户请求
type CreateBizUserByAccountRequest struct {
	TenantID int64  `json:"tenant_id,string"`
	Phone    string `json:"phone"`    // 手机号（与 Email 二选一）
	Email    string `json:"email"`    // 邮箱（与 Phone 二选一）
	Nickname string `json:"nickname"`
}

// UpdateBizUserRequest 更新C端用户请求
type UpdateBizUserRequest struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Version  int    `json:"version" binding:"required"`
	// Token 过期时间（分钟），NULL 时使用全局配置
	AccessTokenExpireMinutes  *int `json:"access_token_expire_minutes"`
	RefreshTokenExpireMinutes *int `json:"refresh_token_expire_minutes"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Password string `json:"password"`
}
