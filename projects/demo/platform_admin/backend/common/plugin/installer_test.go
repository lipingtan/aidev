package plugin

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go-admin/app/plugin/models"
	"go-admin/common/auth/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建 SQLite 内存数据库并自动迁移模型
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.SysPlugin{}, &model.Application{}); err != nil {
		t.Fatalf("迁移表结构失败: %v", err)
	}
	return db
}

// createTestZip 创建包含 plugin.json 和模拟二进制的 zip 包
func createTestZip(t *testing.T, manifest ManifestV2) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// 写入 plugin.json
	pjData, _ := json.Marshal(manifest)
	f, err := w.Create("plugin.json")
	if err != nil {
		t.Fatalf("创建 zip 条目失败: %v", err)
	}
	if _, err := f.Write(pjData); err != nil {
		t.Fatalf("写入 zip 条目失败: %v", err)
	}

	// 写入模拟二进制文件
	binName := manifest.Name
	bf, err := w.Create(binName)
	if err != nil {
		t.Fatalf("创建二进制 zip 条目失败: %v", err)
	}
	bf.Write([]byte("fake-binary"))

	if err := w.Close(); err != nil {
		t.Fatalf("关闭 zip 写入器失败: %v", err)
	}
	return buf
}

// newTestInstaller 创建测试用 Installer 实例
func newTestInstaller(t *testing.T, db *gorm.DB) (*Installer, string, string) {
	t.Helper()
	pluginsDir := t.TempDir()
	staticDir := t.TempDir()
	syncer := NewPluginResourceSyncer(db)
	ins := NewInstaller(pluginsDir, staticDir, db, syncer)
	return ins, pluginsDir, staticDir
}

// TestInstallFromFile_Success 测试正常安装流程
func TestInstallFromFile_Success(t *testing.T) {
	db := setupTestDB(t)
	ins, pluginsDir, _ := newTestInstaller(t, db)

	manifest := ManifestV2{
		Name:        "test-plugin",
		Version:     "1.0.0",
		DisplayName: "测试插件",
		Description: "这是一个测试插件",
		RoutePrefix: "/api/v1/test",
		Platforms:   []string{"admin:pc"},
		Modules:     []ManifestModule{{Code: "mod1", Name: "模块1"}},
	}
	zipBuf := createTestZip(t, manifest)

	err := ins.InstallFromFile(context.Background(), "", zipBuf, "test-plugin.zip")
	if err != nil {
		t.Fatalf("安装应成功: %v", err)
	}

	// 验证 sys_plugin 记录
	var record models.SysPlugin
	if err := db.Where("name = ?", "test-plugin").First(&record).Error; err != nil {
		t.Fatalf("应存在 sys_plugin 记录: %v", err)
	}
	if record.Version != "1.0.0" {
		t.Errorf("版本应为 1.0.0，实际: %s", record.Version)
	}

	// 验证 admin_application 记录
	var app model.Application
	if err := db.Where("app_code = ?", "test-plugin").First(&app).Error; err != nil {
		t.Fatalf("应存在 admin_application 记录: %v", err)
	}
	if app.AppType != "PLUGIN" {
		t.Errorf("app_type 应为 PLUGIN，实际: %s", app.AppType)
	}
	if app.Name != "测试插件" {
		t.Errorf("name 应为 '测试插件'，实际: %s", app.Name)
	}

	// 验证插件目录存在
	destDir := filepath.Join(pluginsDir, "test-plugin")
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		t.Error("插件目录应存在")
	}
}

// TestInstallFromFile_DuplicatePlugin 测试重复安装检测
func TestInstallFromFile_DuplicatePlugin(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	manifest := ManifestV2{Name: "dup-plugin", Version: "1.0.0", DisplayName: "重复插件"}
	zipBuf := createTestZip(t, manifest)

	// 第一次安装
	err := ins.InstallFromFile(context.Background(), "", zipBuf, "dup-plugin.zip")
	if err != nil {
		t.Fatalf("首次安装应成功: %v", err)
	}

	// 第二次安装应失败
	zipBuf2 := createTestZip(t, manifest)
	err = ins.InstallFromFile(context.Background(), "", zipBuf2, "dup-plugin.zip")
	if err == nil {
		t.Fatal("重复安装应返回错误")
	}
	if !contains(err.Error(), "已安装") {
		t.Errorf("错误信息应包含'已安装'，实际: %s", err.Error())
	}
}

// TestInstallFromFile_AppConflict 测试与已有应用冲突检测
func TestInstallFromFile_AppConflict(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	// 预先创建一个同名 Application（模拟内置应用）
	existingApp := &model.Application{
		AppCode: "conflict-app",
		Name:    "已有应用",
		AppType: "BUILTIN",
		Status:  1,
	}
	db.Create(existingApp)

	manifest := ManifestV2{Name: "conflict-app", Version: "1.0.0", DisplayName: "冲突插件"}
	zipBuf := createTestZip(t, manifest)

	err := ins.InstallFromFile(context.Background(), "", zipBuf, "conflict-app.zip")
	if err == nil {
		t.Fatal("与已有应用冲突时应返回错误")
	}
	if !contains(err.Error(), "冲突") {
		t.Errorf("错误信息应包含'冲突'，实际: %s", err.Error())
	}
}

// TestUpgrade_Success 测试正常升级流程
func TestUpgrade_Success(t *testing.T) {
	db := setupTestDB(t)
	ins, pluginsDir, _ := newTestInstaller(t, db)

	// 先安装 v1.0.0
	manifest := ManifestV2{Name: "upgrade-plugin", Version: "1.0.0", DisplayName: "升级测试"}
	zipBuf := createTestZip(t, manifest)
	if err := ins.InstallFromFile(context.Background(), "", zipBuf, "upgrade-plugin.zip"); err != nil {
		t.Fatalf("安装 v1 失败: %v", err)
	}

	// 确保目标目录存在（InstallFromFile 已创建）
	destDir := filepath.Join(pluginsDir, "upgrade-plugin")
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		t.Fatalf("插件目录应存在: %s", destDir)
	}

	// 升级到 v2.0.0
	newManifest := ManifestV2{Name: "upgrade-plugin", Version: "2.0.0", DisplayName: "升级测试V2"}
	newZip := createTestZip(t, newManifest)

	result, err := ins.Upgrade(context.Background(), "upgrade-plugin", newZip, "upgrade-plugin.zip", false)
	if err != nil {
		t.Fatalf("升级应成功: %v", err)
	}
	if result.OldVersion != "1.0.0" {
		t.Errorf("旧版本应为 1.0.0，实际: %s", result.OldVersion)
	}
	if result.NewVersion != "2.0.0" {
		t.Errorf("新版本应为 2.0.0，实际: %s", result.NewVersion)
	}
	if result.NeedConfirm {
		t.Error("非 breaking 升级不应需要确认")
	}

	// 验证数据库更新
	var record models.SysPlugin
	db.Where("name = ?", "upgrade-plugin").First(&record)
	if record.Version != "2.0.0" {
		t.Errorf("sys_plugin version 应为 2.0.0，实际: %s", record.Version)
	}
}

// TestUpgrade_NameMismatch 测试升级包名称不一致
func TestUpgrade_NameMismatch(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	// 安装 original
	manifest := ManifestV2{Name: "original", Version: "1.0.0", DisplayName: "原始"}
	zipBuf := createTestZip(t, manifest)
	if err := ins.InstallFromFile(context.Background(), "", zipBuf, "original.zip"); err != nil {
		t.Fatalf("安装失败: %v", err)
	}

	// 用不同名称的升级包尝试升级
	badManifest := ManifestV2{Name: "different-name", Version: "2.0.0", DisplayName: "不同名"}
	badZip := createTestZip(t, badManifest)

	_, err := ins.Upgrade(context.Background(), "original", badZip, "different.zip", false)
	if err == nil {
		t.Fatal("名称不一致应返回错误")
	}
	if !contains(err.Error(), "不一致") {
		t.Errorf("错误信息应包含'不一致'，实际: %s", err.Error())
	}
}

// TestUpgrade_MinUpgradeFrom 测试 minUpgradeFrom 校验
func TestUpgrade_MinUpgradeFrom(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	// 安装 v1.0.0
	manifest := ManifestV2{Name: "min-ver-plugin", Version: "1.0.0", DisplayName: "最低版本测试"}
	zipBuf := createTestZip(t, manifest)
	if err := ins.InstallFromFile(context.Background(), "", zipBuf, "min-ver-plugin.zip"); err != nil {
		t.Fatalf("安装失败: %v", err)
	}

	// 升级包要求最低从 2.0.0 开始升级
	newManifest := ManifestV2{
		Name:           "min-ver-plugin",
		Version:        "3.0.0",
		DisplayName:    "最低版本测试V3",
		MinUpgradeFrom: "2.0.0",
	}
	newZip := createTestZip(t, newManifest)

	_, err := ins.Upgrade(context.Background(), "min-ver-plugin", newZip, "plugin.zip", false)
	if err == nil {
		t.Fatal("版本过低时应返回错误")
	}
	if !contains(err.Error(), "过低") {
		t.Errorf("错误信息应包含'过低'，实际: %s", err.Error())
	}
}

// TestUpgrade_BreakingUpgrade_NeedConfirm 测试 breaking upgrade 需要确认
func TestUpgrade_BreakingUpgrade_NeedConfirm(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	// 安装 v1.0.0
	manifest := ManifestV2{Name: "breaking-plugin", Version: "1.0.0", DisplayName: "破坏性升级测试"}
	zipBuf := createTestZip(t, manifest)
	if err := ins.InstallFromFile(context.Background(), "", zipBuf, "breaking-plugin.zip"); err != nil {
		t.Fatalf("安装失败: %v", err)
	}

	// breaking upgrade，不传 confirm
	newManifest := ManifestV2{
		Name:            "breaking-plugin",
		Version:         "2.0.0",
		DisplayName:     "破坏性V2",
		BreakingUpgrade: true,
		MigrationNotes:  "需要手动迁移数据",
	}
	newZip := createTestZip(t, newManifest)

	result, err := ins.Upgrade(context.Background(), "breaking-plugin", newZip, "plugin.zip", false)
	if err != nil {
		t.Fatalf("breaking upgrade 无 confirm 应返回结果而非错误: %v", err)
	}
	if !result.NeedConfirm {
		t.Error("应返回 NeedConfirm=true")
	}
	if result.MigrationNotes != "需要手动迁移数据" {
		t.Errorf("MigrationNotes 应为 '需要手动迁移数据'，实际: %s", result.MigrationNotes)
	}

	// 传入 confirm=true 后应正常升级
	newZip2 := createTestZip(t, newManifest)
	result2, err := ins.Upgrade(context.Background(), "breaking-plugin", newZip2, "plugin.zip", true)
	if err != nil {
		t.Fatalf("confirm=true 时升级应成功: %v", err)
	}
	if result2.NeedConfirm {
		t.Error("confirm 后 NeedConfirm 应为 false")
	}
	if result2.NewVersion != "2.0.0" {
		t.Errorf("新版本应为 2.0.0，实际: %s", result2.NewVersion)
	}
}

// TestUpgrade_PluginNotInstalled 测试升级不存在的插件
func TestUpgrade_PluginNotInstalled(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	manifest := ManifestV2{Name: "nonexist", Version: "2.0.0", DisplayName: "不存在"}
	zipBuf := createTestZip(t, manifest)

	_, err := ins.Upgrade(context.Background(), "nonexist", zipBuf, "plugin.zip", false)
	if err == nil {
		t.Fatal("升级不存在的插件应返回错误")
	}
	if !contains(err.Error(), "未安装") {
		t.Errorf("错误信息应包含'未安装'，实际: %s", err.Error())
	}
}

// TestRollback_Success 测试回滚成功
func TestRollback_Success(t *testing.T) {
	db := setupTestDB(t)
	ins, pluginsDir, _ := newTestInstaller(t, db)

	// 手动创建 .bak 目录和正式目录
	name := "rollback-plugin"
	destDir := filepath.Join(pluginsDir, name)
	bakDir := destDir + ".bak"

	os.MkdirAll(destDir, 0755)
	os.MkdirAll(bakDir, 0755)
	// 在 bak 中写入标记文件
	os.WriteFile(filepath.Join(bakDir, "marker.txt"), []byte("old-version"), 0644)

	err := ins.Rollback(name)
	if err != nil {
		t.Fatalf("回滚应成功: %v", err)
	}

	// 验证 bak 已恢复为正式目录
	if _, err := os.Stat(bakDir); !os.IsNotExist(err) {
		t.Error(".bak 目录回滚后应不存在")
	}
	markerPath := filepath.Join(destDir, "marker.txt")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("marker 文件应存在: %v", err)
	}
	if string(data) != "old-version" {
		t.Errorf("marker 内容应为 'old-version'，实际: %s", string(data))
	}
}

// TestRollback_NoBak 测试无备份时回滚
func TestRollback_NoBak(t *testing.T) {
	db := setupTestDB(t)
	ins, _, _ := newTestInstaller(t, db)

	err := ins.Rollback("no-bak-plugin")
	if err == nil {
		t.Fatal("无备份时回滚应返回错误")
	}
	if !contains(err.Error(), "无可回滚") {
		t.Errorf("错误信息应包含'无可回滚'，实际: %s", err.Error())
	}
}

// TestRecoverStaleUpgrade 测试启动时清理残留目录
func TestRecoverStaleUpgrade(t *testing.T) {
	db := setupTestDB(t)
	ins, pluginsDir, _ := newTestInstaller(t, db)

	// 创建残留 .new 目录（应被清理）
	newDir := filepath.Join(pluginsDir, "some-plugin.new")
	os.MkdirAll(newDir, 0755)

	// 创建残留 .bak 目录（正式目录不存在，应恢复）
	bakDir := filepath.Join(pluginsDir, "orphan-plugin.bak")
	os.MkdirAll(bakDir, 0755)
	os.WriteFile(filepath.Join(bakDir, "test.txt"), []byte("data"), 0644)

	// 创建残留 .bak 目录（正式目录存在，应清理 .bak）
	formalDir := filepath.Join(pluginsDir, "existing-plugin")
	os.MkdirAll(formalDir, 0755)
	extraBak := filepath.Join(pluginsDir, "existing-plugin.bak")
	os.MkdirAll(extraBak, 0755)

	err := ins.RecoverStaleUpgrade()
	if err != nil {
		t.Fatalf("RecoverStaleUpgrade 应成功: %v", err)
	}

	// .new 应被清理
	if _, err := os.Stat(newDir); !os.IsNotExist(err) {
		t.Error(".new 目录应被清理")
	}

	// orphan-plugin.bak 应恢复为 orphan-plugin
	restoredDir := filepath.Join(pluginsDir, "orphan-plugin")
	if _, err := os.Stat(restoredDir); os.IsNotExist(err) {
		t.Error("orphan-plugin 应从 .bak 恢复")
	}
	if _, err := os.Stat(bakDir); !os.IsNotExist(err) {
		t.Error("orphan-plugin.bak 恢复后应不存在")
	}

	// existing-plugin.bak 应被清理（正式目录存在）
	if _, err := os.Stat(extraBak); !os.IsNotExist(err) {
		t.Error("existing-plugin.bak 应被清理（正式目录存在）")
	}
	// existing-plugin 正式目录应保留
	if _, err := os.Stat(formalDir); os.IsNotExist(err) {
		t.Error("existing-plugin 正式目录应保留")
	}
}

// TestCompareVersions 测试版本比较函数
func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.3.0", "1.2.9", 1},
		{"v1.0.0", "1.0.0", 0},
		{"1.0", "1.0.0", 0},
		{"1.0.0", "1.0", 0},
		{"0.9.9", "1.0.0", -1},
	}
	for _, tc := range cases {
		got := compareVersions(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// TestUninstall_Success 测试正常卸载流程
func TestUninstall_Success(t *testing.T) {
	db := setupTestDB(t)
	// 额外迁移 syncer 需要的表（简化版，不含完整 RBAC 表）
	db.Exec("CREATE TABLE IF NOT EXISTS admin_resource (id INTEGER PRIMARY KEY, app_code TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_api_permission (id INTEGER PRIMARY KEY, app_code TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_role_resource (id INTEGER PRIMARY KEY, resource_id INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_role_api (id INTEGER PRIMARY KEY, api_permission_id INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_role_app (id INTEGER PRIMARY KEY, app_code TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS admin_tenant_app (id INTEGER PRIMARY KEY, app_code TEXT)")

	ins, pluginsDir, staticDir := newTestInstaller(t, db)

	// 先安装
	manifest := ManifestV2{Name: "uninstall-test", Version: "1.0.0", DisplayName: "卸载测试"}
	zipBuf := createTestZip(t, manifest)
	if err := ins.InstallFromFile(context.Background(), "", zipBuf, "uninstall-test.zip"); err != nil {
		t.Fatalf("安装失败: %v", err)
	}

	// 创建模拟前端目录
	staticPluginDir := filepath.Join(staticDir, "uninstall-test")
	os.MkdirAll(staticPluginDir, 0755)

	// 卸载
	err := ins.Uninstall(context.Background(), "uninstall-test", false)
	if err != nil {
		t.Fatalf("卸载应成功: %v", err)
	}

	// 验证 sys_plugin 记录已删除
	var count int64
	db.Model(&models.SysPlugin{}).Where("name = ?", "uninstall-test").Count(&count)
	if count != 0 {
		t.Error("sys_plugin 记录应被删除")
	}

	// 验证插件目录已删除
	pluginDir := filepath.Join(pluginsDir, "uninstall-test")
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("插件目录应被删除")
	}

	// 验证前端目录已删除
	if _, err := os.Stat(staticPluginDir); !os.IsNotExist(err) {
		t.Error("前端目录应被删除")
	}
}

// contains 辅助：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
