package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
)

// TestDynamicPermission_UserPoolUser_SkipRBAC 验证 UserPool="user" 的请求跳过权限检查直接放行
func TestDynamicPermission_UserPoolUser_SkipRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// 模拟：注入 AuthContext(UserPool=user)，然后调用 DynamicPermissionMiddleware 内部逻辑
	// 由于 DynamicPermissionMiddleware 依赖 DB，这里直接测试 UserPool 判断逻辑的行为
	router.GET("/api/v1/user/profile", func(c *gin.Context) {
		// 模拟 AuthMiddleware 注入 AuthContext（C端用户）
		SetAuthContext(c, &AuthContext{
			UserID:   200,
			TenantID: 1,
			Roles:    []int64{},
			UserPool: strategy.UserPoolUser,
		})
		c.Next()
	}, userPoolCheckMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/profile", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 UserPool=user 直接放行得到 200，实际 %d，body: %s", w.Code, w.Body.String())
	}
}

// TestDynamicPermission_UserPoolAdmin_NeedRBAC 验证 UserPool="admin" 的请求走原有权限逻辑
func TestDynamicPermission_UserPoolAdmin_NeedRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/v1/admin/users", func(c *gin.Context) {
		// 模拟 AuthMiddleware 注入 AuthContext（管理端用户）
		SetAuthContext(c, &AuthContext{
			UserID:   100,
			TenantID: 1,
			Roles:    []int64{1},
			UserPool: strategy.UserPoolAdmin,
		})
		c.Next()
	}, userPoolCheckMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/users", nil)
	router.ServeHTTP(w, req)

	// UserPool=admin 不跳过，走权限逻辑 → 被模拟中间件拒绝（返回 403）
	if w.Code != http.StatusForbidden {
		t.Errorf("期望 UserPool=admin 走权限检查得到 403，实际 %d，body: %s", w.Code, w.Body.String())
	}
}

// TestDynamicPermission_UserPoolEmpty_NeedRBAC 验证旧 token（UserPool=""）走原有权限逻辑
func TestDynamicPermission_UserPoolEmpty_NeedRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/v1/admin/settings", func(c *gin.Context) {
		// 模拟旧 token 场景：UserPool 为空字符串
		SetAuthContext(c, &AuthContext{
			UserID:   100,
			TenantID: 1,
			Roles:    []int64{1},
			UserPool: "",
		})
		c.Next()
	}, userPoolCheckMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/settings", nil)
	router.ServeHTTP(w, req)

	// UserPool="" 不跳过，走权限逻辑 → 被模拟中间件拒绝（返回 403）
	if w.Code != http.StatusForbidden {
		t.Errorf("期望 UserPool=\"\" 走权限检查得到 403，实际 %d，body: %s", w.Code, w.Body.String())
	}
}

// userPoolCheckMiddleware 模拟 DynamicPermissionMiddleware 中的 UserPool 判断逻辑
// 只提取 UserPool 判断部分，不依赖 DB；未通过的请求返回 403 模拟"权限检查失败"
func userPoolCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 40101, "data": nil, "message": "未认证",
			})
			return
		}

		// C端用户（pool=user）跳过 RBAC 权限检查
		if authCtx.UserPool == strategy.UserPoolUser {
			c.Next()
			return
		}

		// 模拟：非 user pool 走权限检查，这里直接拒绝以验证流程
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code": 40301, "data": nil, "message": "权限不足（模拟）",
		})
	}
}
