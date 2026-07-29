package middleware

// highRiskPaths 高风险操作白名单（method:fullPath 格式）
var highRiskPaths = map[string]bool{
	"DELETE:/api/v1/admin/tenants/:id":      true, // 删除租户
	"PUT:/api/v1/admin/users/:id/status":    true, // 禁用/启用用户
	"PUT:/api/v1/admin/users/:id/reset-pwd": true, // 重置密码
	"DELETE:/api/v1/admin/roles/:id":        true, // 删除角色
	"PUT:/api/v1/admin/roles/:id/resources": true, // 修改角色权限（含级联裁剪）
	"PUT:/api/v1/admin/roles/:id/apis":      true, // 修改角色API权限
	"DELETE:/api/v1/admin/applications/:id": true, // 删除应用
}

// GetRiskLevel 根据请求方法和路径判断风险等级
// 返回 "HIGH" / "LOW" / ""（GET 请求不记录操作日志）
func GetRiskLevel(method, fullPath string) string {
	if method == "GET" {
		return ""
	}
	key := method + ":" + fullPath
	if highRiskPaths[key] {
		return "HIGH"
	}
	return "LOW"
}
