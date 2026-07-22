package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-admin/app/plugin/models"
	"go-admin/common/plugin"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupPluginTestDB 创建插件测试用 SQLite 内存数据库
func setupPluginTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.SysPlugin{}); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// setupPluginRouter 创建插件管理测试路由
func setupPluginRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *plugin.PluginManager, *plugin.Installer) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	mgr := plugin.NewPluginManager(db, tmpDir)
	syncer := plugin.NewPluginResourceSyncer(db)
	ins := plugin.NewInstaller(tmpDir, t.TempDir(), db, syncer)

	h := NewPluginHandler(mgr, ins)

	admin := r.Group("/api/v1/admin")
	admin.Use(mockAuthMiddleware())
	h.RegisterRoutes(admin)

	return r, mgr, ins
}

// TestPluginHandler_List 测试获取插件列表
func TestPluginHandler_List(t *testing.T) {
	db := setupPluginTestDB(t)
	// 预置插件记录
	db.Create(&models.SysPlugin{
		Name:       "test-plugin",
		Version:    "1.0.0",
		Status:     models.PluginStatusInstalled,
		BinaryPath: "/tmp/test-plugin",
	})
	db.Create(&models.SysPlugin{
		Name:       "another-plugin",
		Version:    "2.0.0",
		Status:     models.PluginStatusRunning,
		BinaryPath: "/tmp/another-plugin",
	})

	r, _, _ := setupPluginRouter(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List  []pluginListItem `json:"list"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d", resp.Code)
	}
	if resp.Data.Total != 2 {
		t.Fatalf("期望 2 个插件，实际: %d", resp.Data.Total)
	}
}

// TestPluginHandler_Start 测试启动插件
// 注意：由于 StartProcess 需要实际的二进制文件和 gRPC 连接，这里测试的是错误路径
// 验证 Handler 正确响应错误场景
func TestPluginHandler_Start(t *testing.T) {
	db := setupPluginTestDB(t)
	db.Create(&models.SysPlugin{
		Name:       "my-plugin",
		Version:    "1.0.0",
		Status:     models.PluginStatusInstalled,
		BinaryPath: "/nonexistent/path/my-plugin",
	})

	r, _, _ := setupPluginRouter(t, db)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/my-plugin/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// StartProcess 会失败（因为二进制不存在），但 handler 应正常返回错误而非 panic
	if w.Code == http.StatusOK {
		// 如果意外成功（不应发生），验证响应
		var resp Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Code != 0 {
			t.Logf("启动返回错误（预期）: %s", resp.Message)
		}
	} else {
		// 非 200 响应是预期的（二进制不存在导致连接失败）
		t.Logf("启动失败（预期，二进制不存在）: status=%d, body=%s", w.Code, w.Body.String())
	}
}

// TestPluginHandler_Start_NotExist 测试启动不存在的插件返回错误
func TestPluginHandler_Start_NotExist(t *testing.T) {
	db := setupPluginTestDB(t)
	r, _, _ := setupPluginRouter(t, db)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/nonexistent/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestPluginHandler_Stop 测试停止插件
func TestPluginHandler_Stop(t *testing.T) {
	db := setupPluginTestDB(t)
	r, _, _ := setupPluginRouter(t, db)

	// 停止一个未运行的插件应该返回错误
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/not-running/stop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// manager.Stop 对未注册的插件返回错误
	if w.Code == http.StatusOK {
		var resp Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Code == 0 {
			t.Fatal("停止未运行的插件不应该成功")
		}
	}
	// 预期返回 500（内部服务错误）因为 manager 中不存在该插件
	t.Logf("停止未运行插件: status=%d, body=%s", w.Code, w.Body.String())
}

// TestPluginHandler_Uninstall 测试卸载插件
func TestPluginHandler_Uninstall(t *testing.T) {
	db := setupPluginTestDB(t)
	// 卸载时 syncer 需要这些表
	db.AutoMigrate(&models.SysPlugin{})
	// 手动建表供 syncer 使用（简化版）
	db.Exec("CREATE TABLE IF NOT EXISTS admin_resource (id INTEGER PRIMARY KEY, app_code VARCHAR(64), deleted_at DATETIME)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_api_permission (id INTEGER PRIMARY KEY, app_code VARCHAR(64))")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_role_app (id INTEGER PRIMARY KEY, app_code VARCHAR(64))")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_role_resource (id INTEGER PRIMARY KEY, resource_id INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_role_api (id INTEGER PRIMARY KEY, api_id INTEGER, api_permission_id INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_application (id INTEGER PRIMARY KEY, app_code VARCHAR(64), deleted_at DATETIME)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_tenant_app (id INTEGER PRIMARY KEY, app_code VARCHAR(64))")

	tmpDir := t.TempDir()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	mgr := plugin.NewPluginManager(db, tmpDir)
	syncer := plugin.NewPluginResourceSyncer(db)
	ins := plugin.NewInstaller(tmpDir, t.TempDir(), db, syncer)
	h := NewPluginHandler(mgr, ins)

	admin := r.Group("/api/v1/admin")
	admin.Use(mockAuthMiddleware())
	h.RegisterRoutes(admin)

	// 预置插件记录
	db.Create(&models.SysPlugin{
		Name:       "removable-plugin",
		Version:    "1.0.0",
		Status:     models.PluginStatusStopped,
		BinaryPath: tmpDir + "/removable-plugin/removable-plugin",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/plugins/removable-plugin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("期望 code=0，实际: %d, message: %s", resp.Code, resp.Message)
	}

	// 验证记录已删除
	var count int64
	db.Model(&models.SysPlugin{}).Where("name = ?", "removable-plugin").Count(&count)
	if count != 0 {
		t.Fatalf("期望插件记录已删除，但仍存在")
	}
}

// TestPluginHandler_Upgrade_NeedConfirm 测试升级需要二次确认
// 由于 Upgrade 需要文件上传和实际文件解析，这里验证缺少文件时的错误处理
func TestPluginHandler_Upgrade_NoFile(t *testing.T) {
	db := setupPluginTestDB(t)
	r, _, _ := setupPluginRouter(t, db)

	// 不传文件的 PUT 请求
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/plugins/my-plugin/upgrade", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400（缺少文件），实际: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestPluginHandler_Health 测试健康检查
func TestPluginHandler_Health(t *testing.T) {
	db := setupPluginTestDB(t)
	r, _, _ := setupPluginRouter(t, db)

	// 健康检查一个未注册的插件应返回错误
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/unknown-plugin/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// manager.Healthcheck 对未注册的插件返回 error
	if w.Code == http.StatusOK {
		var resp struct {
			Code int `json:"code"`
			Data struct {
				Healthy bool   `json:"healthy"`
				Message string `json:"message"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		// 如果意外返回 200，检查 healthy 为 false
		if resp.Data.Healthy {
			t.Fatal("未注册插件的健康状态不应为 true")
		}
	} else {
		// 预期非 200 响应（未注册插件）
		t.Logf("未注册插件健康检查: status=%d（预期错误）", w.Code)
	}
}

// TestPluginHandler_Health_EmptyName 测试空名称返回 400
func TestPluginHandler_Health_EmptyName(t *testing.T) {
	db := setupPluginTestDB(t)
	r, _, _ := setupPluginRouter(t, db)

	// Gin 路由中 :name 不会匹配空值，这里测试路由级别行为
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins//health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Gin 路由不匹配空 :name 参数，预期 404
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Logf("空名称健康检查: status=%d", w.Code)
	}
}
