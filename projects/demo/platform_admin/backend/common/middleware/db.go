package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk"
)

func WithContextDb(c *gin.Context) {
	// 未安装时 db 为 nil，跳过注入，避免 panic
	db := sdk.Runtime.GetDbByKey(c.Request.Host)
	if db != nil {
		c.Set("db", db.WithContext(c))
	}
	c.Next()
}
