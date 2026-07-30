package actions

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DataPermission 数据权限结构（V2 已用 DataScopeCallback 替代，保留编译兼容）
type DataPermission struct {
	DataScope string
	UserId    int
	DeptId    int
	RoleId    int
}

// GetPermissionFromContext 从 gin.Context 获取数据权限（V2 no-op，返回空权限）
func GetPermissionFromContext(c *gin.Context) *DataPermission {
	return &DataPermission{}
}

// Permission 数据权限 GORM Scope（V2 no-op，直接返回原 db）
func Permission(tableName string, p *DataPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db
	}
}

// PermissionAction 数据权限中间件（V2 no-op）
func PermissionAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
