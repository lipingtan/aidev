package dto

// CreateBizUserRequest 创建C端用户请求
// TenantID 由 handler 从 AuthContext 自动注入，前端无需传
type CreateBizUserRequest struct {
	TenantID int64  `json:"tenant_id,string"` // handler 层从 AuthContext 注入，binding 不强制
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password"`  // 可选
	Nickname string `json:"nickname"`
}

// UpdateBizUserRequest 更新C端用户请求
type UpdateBizUserRequest struct {
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Version  int    `json:"version" binding:"required"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Password string `json:"password"`
}
