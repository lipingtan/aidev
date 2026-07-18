package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-admin/common/auth/cache"
	"go-admin/common/auth/config"

	"github.com/gin-gonic/gin"
)

// 构造测试用的 PermissionMiddleware
func setupTestMiddleware(superAdminRoleID int64) (*PermissionMiddleware, *cache.PermissionCache) {
	lc := cache.NewLocalCache()
	tlc := cache.NewTwoLevelCache(lc, lc)
	pc := cache.NewPermissionCache(tlc, 10*time.Minute, 5*time.Minute)

	cfg := config.DefaultConfig()

	checker := func(roleIDs []int64) bool {
		for _, id := range roleIDs {
			if id == superAdminRoleID {
				return true
			}
		}
		return false
	}

	pm := NewPermissionMiddleware(pc, cfg, checker)
	return pm, pc
}

// 设置用户权限到缓存
func setUserPermsInCache(pc *cache.PermissionCache, roleID string, perms []interface{}) {
	pc.SetRolePermissions(roleID, perms)
}

func TestRequirePermission_HasPermission_Pass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pm, pc := setupTestMiddleware(999)

	// 设置角色权限到 L1 缓存
	setUserPermsInCache(pc, "1", []interface{}{"user:list", "user:create"})

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		// 模拟 AuthMiddleware 注入 AuthContext
		SetAuthContext(c, &AuthContext{
			UserID:   100,
			TenantID: 1,
			Roles:    []int64{1},
		})
		c.Next()
	}, pm.RequirePermission("user:list"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200，实际 %d，body: %s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_NoPermission_403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pm, pc := setupTestMiddleware(999)

	// 设置角色权限，不包含目标权限
	setUserPermsInCache(pc, "1", []interface{}{"user:list"})

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		SetAuthContext(c, &AuthContext{
			UserID:   100,
			TenantID: 1,
			Roles:    []int64{1},
		})
		c.Next()
	}, pm.RequirePermission("user:delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("期望 403，实际 %d，body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if code, ok := resp["code"].(float64); !ok || int(code) != 40301 {
		t.Errorf("期望 code=40301，实际 %v", resp["code"])
	}
}

func TestRequirePermission_SuperAdmin_AlwaysPass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 角色 ID 999 为 SUPER_ADMIN
	pm, _ := setupTestMiddleware(999)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		SetAuthContext(c, &AuthContext{
			UserID:   100,
			TenantID: 1,
			Roles:    []int64{999}, // SUPER_ADMIN 角色
		})
		c.Next()
	}, pm.RequirePermission("any:permission:code"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 SUPER_ADMIN 直接通过得到 200，实际 %d，body: %s", w.Code, w.Body.String())
	}
}

func TestRequirePermission_NoAuthContext_401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pm, _ := setupTestMiddleware(999)

	router := gin.New()
	// 不注入 AuthContext，模拟未认证
	router.GET("/test", pm.RequirePermission("user:list"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401，实际 %d，body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if code, ok := resp["code"].(float64); !ok || int(code) != 40101 {
		t.Errorf("期望 code=40101，实际 %v", resp["code"])
	}
}
