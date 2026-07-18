package middleware

import (
	"net/http"
	"strings"

	"go-admin/common/auth/service"

	"github.com/gin-gonic/gin"
)

const authContextKey = "auth_context"

// AuthContext 认证上下文，注入到 gin.Context
type AuthContext struct {
	UserID   int64
	TenantID int64
	Roles    []int64
}

// SetAuthContext 设置认证上下文到 gin.Context
func SetAuthContext(c *gin.Context, ctx *AuthContext) {
	c.Set(authContextKey, ctx)
}

// GetAuthContext 从 gin.Context 获取认证上下文
func GetAuthContext(c *gin.Context) *AuthContext {
	val, exists := c.Get(authContextKey)
	if !exists {
		return nil
	}
	ctx, ok := val.(*AuthContext)
	if !ok {
		return nil
	}
	return ctx
}

// AuthMiddleware 认证中间件
// 解析 access_token、检查黑名单、注入 AuthContext
func AuthMiddleware(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取 Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"data":    nil,
				"message": "缺少认证信息",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"data":    nil,
				"message": "认证格式错误",
			})
			return
		}

		tokenStr := parts[1]

		// 解析 access_token
		claims, err := authSvc.ParseAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40102,
				"data":    nil,
				"message": "token 无效或已过期",
			})
			return
		}

		// 检查黑名单
		if authSvc.IsBlacklisted(claims.ID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40103,
				"data":    nil,
				"message": "token 已失效",
			})
			return
		}

		// 检查租户状态
		tenantStatus, err := authSvc.CheckTenantStatus(claims.TenantID)
		if err == nil && tenantStatus != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    40105,
				"data":    nil,
				"message": "租户已禁用",
			})
			return
		}

		// 检查用户状态
		userStatus, err := authSvc.CheckUserStatus(claims.UserID)
		if err == nil && userStatus != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40104,
				"data":    nil,
				"message": "账号已禁用",
			})
			return
		}

		// 注入 AuthContext
		authCtx := &AuthContext{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Roles:    claims.Roles,
		}
		SetAuthContext(c, authCtx)

		c.Next()
	}
}
