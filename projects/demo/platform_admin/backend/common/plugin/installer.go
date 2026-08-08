package plugin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go-admin/app/plugin/models"
	authModel "go-admin/common/auth/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Installer 插件安装/卸载管理器
type Installer struct {
	pluginsDir string                // 插件二进制存放目录（如 ./plugins/）
	staticDir  string                // 前端 bundle 存放目录（如 ./static/plugins/）
	db         *gorm.DB              // 数据库连接
	syncer     *PluginResourceSyncer // 资源同步器
}

// NewInstaller 创建安装器实例
func NewInstaller(pluginsDir, staticDir string, db *gorm.DB, syncer *PluginResourceSyncer) *Installer {
	return &Installer{
		pluginsDir: pluginsDir,
		staticDir:  staticDir,
		db:         db,
		syncer:     syncer,
	}
}

// UpgradeResult 升级操作结果
type UpgradeResult struct {
	NeedConfirm    bool   `json:"needConfirm"`    // 是否需要二次确认
	MigrationNotes string `json:"migrationNotes"` // 迁移说明
	OldVersion     string `json:"oldVersion"`     // 旧版本号
	NewVersion     string `json:"newVersion"`     // 新版本号
}

// InstallFromFile 从文件流安装插件
// 自动从 plugin.json 提取插件名，不需要外部传入
// file: 文件内容读取器
// filename: 原始文件名，用于判断压缩格式（.zip 或 .tar.gz）
func (ins *Installer) InstallFromFile(ctx context.Context, _ string, file io.Reader, filename string) error {
	// 先解压到临时目录
	tempDir := filepath.Join(ins.pluginsDir, "_temp_install")
	_ = os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 根据文件后缀选择解压方式
	var err error
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		err = extractTarGz(file, tempDir)
	} else if strings.HasSuffix(filename, ".zip") {
		err = extractZip(file, tempDir)
	} else {
		return fmt.Errorf("不支持的文件格式: %s（仅支持 .zip 和 .tar.gz）", filename)
	}
	if err != nil {
		return fmt.Errorf("解压文件失败: %w", err)
	}

	// 从 plugin.json 读取完整 ManifestV2
	pjPath := filepath.Join(tempDir, "plugin.json")
	pjData, readErr := os.ReadFile(pjPath)
	if readErr != nil {
		return fmt.Errorf("插件包中缺少 plugin.json 文件")
	}
	var manifest ManifestV2
	if jsonErr := json.Unmarshal(pjData, &manifest); jsonErr != nil {
		return fmt.Errorf("plugin.json 格式错误: %w", jsonErr)
	}
	if manifest.Name == "" {
		return fmt.Errorf("plugin.json 中 name 字段不能为空")
	}
	name := manifest.Name

	// 检查是否已安装同名插件
	var existCount int64
	ins.db.Model(&models.SysPlugin{}).Where("name = ?", name).Count(&existCount)
	if existCount > 0 {
		return fmt.Errorf("插件 %s 已安装，请先卸载后再安装", name)
	}

	// admin_application 冲突检查
	var appConflict int64
	ins.db.Model(&authModel.Application{}).Where("app_code = ? AND deleted_at IS NULL", name).Count(&appConflict)
	if appConflict > 0 {
		return fmt.Errorf("插件名称 %s 与已有应用冲突，无法安装", name)
	}

	// 安装前先 kill 可能残留的同名插件进程
	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}
	killPluginProcess(binaryName)
	time.Sleep(500 * time.Millisecond)

	// 移动到正式目录
	destDir := filepath.Join(ins.pluginsDir, name)
	_ = os.RemoveAll(destDir)
	if err := os.Rename(tempDir, destDir); err != nil {
		// Rename 跨分区可能失败，用 copy 兜底
		if cpErr := copyDir(tempDir, destDir); cpErr != nil {
			return fmt.Errorf("移动插件文件失败: %w", cpErr)
		}
	}

	// 查找二进制文件
	binaryPath := filepath.Join(destDir, binaryName)
	if _, statErr := os.Stat(binaryPath); statErr != nil {
		binaryPath = destDir
	}

	// 部署前端 bundle（V2 多端支持）
	var frontendPath string
	if len(manifest.Frontends) > 0 {
		if err := ins.deployFrontendBundles(name, destDir, manifest.Frontends); err != nil {
			// 清理已解压文件
			_ = os.RemoveAll(destDir)
			return fmt.Errorf("部署前端 bundle 失败: %w", err)
		}
		// frontendPath 设为插件静态根目录（兼容性）
		frontendPath = filepath.Join(ins.staticDir, "plugins", name)
	}

	// 写入 sys_plugin 数据库记录
	version := manifest.Version
	if version == "" {
		version = "0.0.1"
	}
	now := time.Now()
	record := &models.SysPlugin{
		Name:         name,
		Version:      version,
		Description:  manifest.Description,
		Status:       models.PluginStatusInstalled,
		BinaryPath:   binaryPath,
		FrontendPath: frontendPath,
		InstalledAt:  &now,
	}
	if err = ins.db.WithContext(ctx).Create(record).Error; err != nil {
		// 数据库写入失败时清理已解压文件
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("写入插件记录失败: %w", err)
	}

	// 创建 admin_application 记录
	app := &authModel.Application{
		AppCode:     name,
		Name:        manifest.DisplayName,
		Description: manifest.Description,
		AppType:     "PLUGIN",
		RoutePrefix: manifest.RoutePrefix,
		Platforms:   marshalJSONBytes(manifest.Platforms),
		Modules:     marshalJSONBytes(manifest.Modules),
		Status:      1,
	}
	if err = ins.db.WithContext(ctx).Create(app).Error; err != nil {
		// 回滚：清理 sys_plugin 记录和文件
		ins.db.WithContext(ctx).Unscoped().Where("name = ?", name).Delete(&models.SysPlugin{})
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("创建应用记录失败: %w", err)
	}

	return nil
}

// Upgrade 升级插件（安全文件替换 + 二次确认机制）
func (ins *Installer) Upgrade(ctx context.Context, name string, file io.Reader, filename string, confirm bool) (*UpgradeResult, error) {
	// 1. 校验目标插件存在
	var record models.SysPlugin
	if err := ins.db.Where("name = ?", name).First(&record).Error; err != nil {
		return nil, fmt.Errorf("插件 %s 未安装", name)
	}

	// 2. 解压到临时目录
	tempDir := filepath.Join(ins.pluginsDir, "_temp_upgrade")
	_ = os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var err error
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		err = extractTarGz(file, tempDir)
	} else if strings.HasSuffix(filename, ".zip") {
		err = extractZip(file, tempDir)
	} else {
		return nil, fmt.Errorf("不支持的文件格式: %s（仅支持 .zip 和 .tar.gz）", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("解压文件失败: %w", err)
	}

	// 3. 解析 manifest
	pjPath := filepath.Join(tempDir, "plugin.json")
	pjData, readErr := os.ReadFile(pjPath)
	if readErr != nil {
		return nil, fmt.Errorf("升级包中缺少 plugin.json 文件")
	}
	var manifest ManifestV2
	if jsonErr := json.Unmarshal(pjData, &manifest); jsonErr != nil {
		return nil, fmt.Errorf("plugin.json 格式错误: %w", jsonErr)
	}

	// 4. 校验 name 一致
	if manifest.Name != name {
		return nil, fmt.Errorf("升级包 name=%s 与目标插件 %s 不一致", manifest.Name, name)
	}

	// 5. minUpgradeFrom 校验
	if manifest.MinUpgradeFrom != "" && compareVersions(record.Version, manifest.MinUpgradeFrom) < 0 {
		return nil, fmt.Errorf("当前版本 %s 过低，请先升级到 %s 再执行此升级", record.Version, manifest.MinUpgradeFrom)
	}

	// 6. breakingUpgrade 检测
	if manifest.BreakingUpgrade && !confirm {
		return &UpgradeResult{
			NeedConfirm:    true,
			MigrationNotes: manifest.MigrationNotes,
			OldVersion:     record.Version,
			NewVersion:     manifest.Version,
		}, nil
	}

	// 7. 安全文件替换（后端）
	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}
	killPluginProcess(binaryName)
	time.Sleep(500 * time.Millisecond)

	destDir := filepath.Join(ins.pluginsDir, name)
	newDir := destDir + ".new"
	bakDir := destDir + ".bak"
	_ = os.RemoveAll(newDir)
	_ = os.RemoveAll(bakDir)

	// rename tempDir → newDir
	if err := os.Rename(tempDir, newDir); err != nil {
		if cpErr := copyDir(tempDir, newDir); cpErr != nil {
			return nil, fmt.Errorf("准备新版本目录失败: %w", cpErr)
		}
	}
	// rename destDir → bakDir
	if err := os.Rename(destDir, bakDir); err != nil {
		_ = os.RemoveAll(newDir)
		return nil, fmt.Errorf("备份旧版本失败: %w", err)
	}
	// rename newDir → destDir
	if err := os.Rename(newDir, destDir); err != nil {
		// 恢复：bakDir → destDir
		_ = os.Rename(bakDir, destDir)
		return nil, fmt.Errorf("安装新版本失败: %w", err)
	}

	// 8. 前端 bundle 同理
	frontendSrc := filepath.Join(destDir, "frontend", "dist")
	if info, statErr := os.Stat(frontendSrc); statErr == nil && info.IsDir() {
		staticDir := filepath.Join(ins.staticDir, name)
		staticNew := staticDir + ".new"
		staticBak := staticDir + ".bak"
		_ = os.RemoveAll(staticNew)
		_ = os.RemoveAll(staticBak)

		if cpErr := copyDir(frontendSrc, staticNew); cpErr == nil {
			if _, statErr := os.Stat(staticDir); statErr == nil {
				_ = os.Rename(staticDir, staticBak)
			}
			_ = os.Rename(staticNew, staticDir)
		}
	}

	// 9. 更新 sys_plugin
	binaryPath := filepath.Join(destDir, binaryName)
	ins.db.Model(&models.SysPlugin{}).Where("name = ?", name).Updates(map[string]interface{}{
		"version":     manifest.Version,
		"binary_path": binaryPath,
	})

	// 10. 更新 admin_application
	platformsJSON, _ := json.Marshal(manifest.Platforms)
	modulesJSON, _ := json.Marshal(manifest.Modules)
	ins.db.Model(&authModel.Application{}).Where("app_code = ?", name).Updates(map[string]interface{}{
		"name":         manifest.DisplayName,
		"description":  manifest.Description,
		"route_prefix": manifest.RoutePrefix,
		"platforms":    datatypes.JSON(platformsJSON),
		"modules":      datatypes.JSON(modulesJSON),
	})

	return &UpgradeResult{
		OldVersion: record.Version,
		NewVersion: manifest.Version,
	}, nil
}

// Rollback 回滚升级（将 .bak 恢复为正式目录）
func (ins *Installer) Rollback(name string) error {
	destDir := filepath.Join(ins.pluginsDir, name)
	bakDir := destDir + ".bak"
	if _, err := os.Stat(bakDir); os.IsNotExist(err) {
		return fmt.Errorf("无可回滚的备份: %s", bakDir)
	}
	_ = os.RemoveAll(destDir)
	if err := os.Rename(bakDir, destDir); err != nil {
		return fmt.Errorf("回滚失败: %w", err)
	}
	// 前端 bundle 同理
	staticDir := filepath.Join(ins.staticDir, name)
	staticBak := staticDir + ".bak"
	if _, err := os.Stat(staticBak); err == nil {
		_ = os.RemoveAll(staticDir)
		_ = os.Rename(staticBak, staticDir)
	}
	return nil
}

// RecoverStaleUpgrade 启动时检测残留 .new/.bak 目录并自动恢复
func (ins *Installer) RecoverStaleUpgrade() error {
	entries, err := os.ReadDir(ins.pluginsDir)
	if err != nil {
		return nil // 目录不存在不报错
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		if strings.HasSuffix(dirName, ".new") {
			// 清理残留的 .new 目录
			os.RemoveAll(filepath.Join(ins.pluginsDir, dirName))
		} else if strings.HasSuffix(dirName, ".bak") {
			// 如果正式目录不存在，从 .bak 恢复
			baseName := strings.TrimSuffix(dirName, ".bak")
			formal := filepath.Join(ins.pluginsDir, baseName)
			if _, err := os.Stat(formal); os.IsNotExist(err) {
				os.Rename(filepath.Join(ins.pluginsDir, dirName), formal)
			} else {
				// 正式目录存在，清理多余 .bak
				os.RemoveAll(filepath.Join(ins.pluginsDir, dirName))
			}
		}
	}
	return nil
}

// InstallFromURL 从远程 URL 下载并安装插件
func (ins *Installer) InstallFromURL(ctx context.Context, name string, url string) error {
	// 下载文件
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建下载请求失败: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载插件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载插件失败，HTTP 状态码: %d", resp.StatusCode)
	}

	// 从 URL 推断文件名
	filename := filepath.Base(url)
	if filename == "" || filename == "." || filename == "/" {
		filename = name + ".zip"
	}

	return ins.InstallFromFile(ctx, name, resp.Body, filename)
}

// deployFrontendBundles 部署插件前端 bundle 到静态目录（V2 多端支持）
// pluginName: 插件名称
// pluginDir: 插件解压目录
// frontends: frontends 配置数组
func (ins *Installer) deployFrontendBundles(pluginName string, pluginDir string, frontends []FrontendConfig) error {
	for _, fe := range frontends {
		// 源文件路径：{pluginDir}/{entry}（entry 直接是相对于插件根的路径）
		src := filepath.Join(pluginDir, fe.Entry)

		// 检查源文件是否存在
		if _, err := os.Stat(src); os.IsNotExist(err) {
			return fmt.Errorf("bundle 源文件不存在: %s (platform=%s, device=%s)", src, fe.Platform, fe.Device)
		}

		platform := strings.ToLower(fe.Platform)
		device := strings.ToLower(fe.Device)

		// V2 路径：{staticDir}/plugins/{name}/{platform}-{device}/bundle.js
		destDir := filepath.Join(ins.staticDir, "plugins", pluginName, platform+"-"+device)
		destFile := filepath.Join(destDir, "bundle.js")
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("创建 bundle 目录失败 %s: %w", destDir, err)
		}
		if err := copyFile(src, destFile); err != nil {
			return fmt.Errorf("复制 bundle 失败 %s → %s: %w", src, destFile, err)
		}

		// 兼容路径：admin-pc bundle 同时部署到 {staticDir}/plugins/{name}/index.js
		// 供前端 plugin-loader.ts 通过 /static/plugins/{name}/index.js 加载
		if platform == "admin" && device == "pc" {
			compatDir := filepath.Join(ins.staticDir, "plugins", pluginName)
			if err := os.MkdirAll(compatDir, 0755); err != nil {
				return fmt.Errorf("创建兼容路径目录失败: %w", err)
			}
			compatFile := filepath.Join(compatDir, "index.js")
			if err := copyFile(src, compatFile); err != nil {
				return fmt.Errorf("复制兼容 bundle 失败: %w", err)
			}
		}
	}
	return nil
}

// removeFrontendBundles 删除插件的所有前端 bundle
func (ins *Installer) removeFrontendBundles(pluginName string) error {
	pluginStaticDir := filepath.Join(ins.staticDir, "plugins", pluginName)
	return os.RemoveAll(pluginStaticDir)
}

// Uninstall 卸载插件
// cleanData: 是否清除插件相关数据表（{name}_* 表）
func (ins *Installer) Uninstall(ctx context.Context, name string, cleanData bool) error {
	// 先检查插件记录是否存在（不存在直接返回 not found，避免无效的进程 kill 操作）
	var count int64
	ins.db.WithContext(ctx).Model(&models.SysPlugin{}).Where("name = ?", name).Count(&count)
	if count == 0 {
		return fmt.Errorf("插件 %s 不存在", name)
	}

	// 尝试 kill 可能残留的插件进程（Windows 上 exe 被占用时无法删除）
	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}
	killPluginProcess(binaryName)

	// 短暂等待进程退出
	time.Sleep(500 * time.Millisecond)

	// 清理插件 RBAC 资源
	if err := ins.syncer.SyncOnUninstall(name); err != nil {
		return fmt.Errorf("清理插件 RBAC 资源失败: %w", err)
	}

	// 删除插件目录
	pluginDir := filepath.Join(ins.pluginsDir, name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("删除插件目录失败: %w", err)
	}

	// 删除前端静态文件目录（V2 结构：{staticDir}/plugins/{name}/）
	staticPluginDir := filepath.Join(ins.staticDir, "plugins", name)
	_ = os.RemoveAll(staticPluginDir)

	// 可选：清除插件数据表
	if cleanData {
		if err := ins.dropPluginTables(ctx, name); err != nil {
			return fmt.Errorf("清除插件数据表失败: %w", err)
		}
	}

	// 删除数据库记录（硬删除）
	if err := ins.db.WithContext(ctx).Unscoped().Where("name = ?", name).Delete(&models.SysPlugin{}).Error; err != nil {
		return fmt.Errorf("删除插件记录失败: %w", err)
	}

	return nil
}

// GetPluginRecord 查询指定插件的数据库记录
func (ins *Installer) GetPluginRecord(name string) (*models.SysPlugin, error) {
	var record models.SysPlugin
	if err := ins.db.Where("name = ?", name).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ListPluginRecords 列出所有插件记录
func (ins *Installer) ListPluginRecords() ([]models.SysPlugin, error) {
	var records []models.SysPlugin
	if err := ins.db.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// UpdateStatus 更新插件状态
func (ins *Installer) UpdateStatus(name string, status int) error {
	result := ins.db.Model(&models.SysPlugin{}).Where("name = ?", name).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("插件 %s 不存在", name)
	}
	return nil
}

// Deprecated: 使用 PluginResourceSyncer.SyncOnStart 替代。保留代码以兼容旧版，不再被主流程调用。
// RegisterMenus 将插件菜单写入 sys_menu 表（挂在"扩展功能"目录下）
func (ins *Installer) RegisterMenus(pluginName string, menus []interface{}) error {
	if len(menus) == 0 {
		return nil
	}
	return nil
}

// ensureExtensionMenu 确保"插件扩展"顶级资源节点存在，返回其 resource_id
// 已迁移到 admin_resource 表（V2 菜单体系）
func (ins *Installer) ensureExtensionMenu() int64 {
	const resourceName = "插件扩展"
	const appCode = "platform_admin"
	var resourceID int64
	ins.db.Table("admin_resource").
		Where("name = ? AND app_code = ? AND deleted_at IS NULL", resourceName, appCode).
		Select("id").Scan(&resourceID)
	if resourceID > 0 {
		return resourceID
	}

	// 创建"插件扩展"顶级菜单节点（admin_resource）
	ins.db.Exec(`INSERT INTO admin_resource (id, name, app_code, resource_type, path, icon, sort_order, is_hidden, created_at, updated_at)
		VALUES (?, ?, ?, 'MENU', '/plugin-extensions', 'Grid', 999, 0, NOW(), NOW())`,
		nextResourceID(), resourceName, appCode,
	)

	ins.db.Table("admin_resource").
		Where("name = ? AND app_code = ? AND deleted_at IS NULL", resourceName, appCode).
		Select("id").Scan(&resourceID)
	return resourceID
}

// nextResourceID 生成雪花 ID（复用已有 model.NextID）
func nextResourceID() int64 {
	return authModel.NextID()
}

// Deprecated: 保留接口兼容，不再写入 sys_menu
// UnregisterMenus 从 admin_resource 中软删除插件菜单
func (ins *Installer) UnregisterMenus(pluginName string) error {
	prefix := "plugin_" + pluginName + "%"
	return ins.db.Exec("UPDATE admin_resource SET deleted_at = NOW() WHERE name LIKE ? AND deleted_at IS NULL", prefix).Error
}

// dropPluginTables 删除以 {name}_ 为前缀的所有数据表
func (ins *Installer) dropPluginTables(ctx context.Context, name string) error {
	prefix := name + "_"
	var tables []string
	// 查询所有表名
	if err := ins.db.WithContext(ctx).Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		return err
	}
	for _, table := range tables {
		if strings.HasPrefix(table, prefix) {
			if err := ins.db.WithContext(ctx).Exec("DROP TABLE IF EXISTS `" + table + "`").Error; err != nil {
				return fmt.Errorf("删除表 %s 失败: %w", table, err)
			}
		}
	}
	return nil
}

// compareVersions 比较两个语义化版本号（简化版：按 major.minor.patch 数字逐段比较）
// 返回 -1, 0, 1
func compareVersions(a, b string) int {
	aParts := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bParts := strings.Split(strings.TrimPrefix(b, "v"), ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aNum, bNum int
		if i < len(aParts) {
			aNum, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bNum, _ = strconv.Atoi(bParts[i])
		}
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
	}
	return 0
}

// marshalJSONBytes 将任意值序列化为 datatypes.JSON
func marshalJSONBytes(v interface{}) datatypes.JSON {
	data, _ := json.Marshal(v)
	return datatypes.JSON(data)
}

// extractZip 解压 zip 文件到目标目录
func extractZip(src io.Reader, dest string) error {
	// zip 需要 io.ReaderAt，先读取全部内容到内存
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("读取 zip 数据失败: %w", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("打开 zip 失败: %w", err)
	}

	for _, f := range reader.File {
		target := filepath.Join(dest, f.Name)

		// 防止 zip slip 攻击
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if mkErr := os.MkdirAll(target, 0755); mkErr != nil {
				return mkErr
			}
			continue
		}

		// 确保父目录存在
		if mkErr := os.MkdirAll(filepath.Dir(target), 0755); mkErr != nil {
			return mkErr
		}

		rc, openErr := f.Open()
		if openErr != nil {
			return openErr
		}

		outFile, createErr := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if createErr != nil {
			rc.Close()
			return createErr
		}

		if _, cpErr := io.Copy(outFile, rc); cpErr != nil {
			outFile.Close()
			rc.Close()
			return cpErr
		}
		outFile.Close()
		rc.Close()
	}
	return nil
}

// extractTarGz 解压 tar.gz 文件到目标目录
func extractTarGz(src io.Reader, dest string) error {
	gz, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("打开 gzip 失败: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, readErr := tr.Next()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}

		target := filepath.Join(dest, header.Name)

		// 防止路径穿越攻击
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if mkErr := os.MkdirAll(target, 0755); mkErr != nil {
				return mkErr
			}
		case tar.TypeReg:
			if mkErr := os.MkdirAll(filepath.Dir(target), 0755); mkErr != nil {
				return mkErr
			}
			outFile, createErr := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if createErr != nil {
				return createErr
			}
			if _, cpErr := io.Copy(outFile, tr); cpErr != nil {
				outFile.Close()
				return cpErr
			}
			outFile.Close()
		}
	}
	return nil
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if cpErr := copyDir(srcPath, dstPath); cpErr != nil {
				return cpErr
			}
		} else {
			if cpErr := copyFile(srcPath, dstPath); cpErr != nil {
				return cpErr
			}
		}
	}
	return nil
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// killPluginProcess 尝试终止插件进程（通过进程名匹配）
func killPluginProcess(binaryName string) {
	if runtime.GOOS == "windows" {
		// Windows: taskkill /F /IM {name}.exe
		_ = exec.Command("taskkill", "/F", "/IM", binaryName).Run()
	} else {
		// Linux/Mac: pkill -f {name}
		_ = exec.Command("pkill", "-f", binaryName).Run()
	}
}
