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
	"strings"
	"time"

	"go-admin/app/plugin/models"
	"game-server/plugin-sdk/proto"
	"gorm.io/gorm"
)

// pluginJSON 插件描述文件结构
type pluginJSON struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// Installer 插件安装/卸载管理器
type Installer struct {
	pluginsDir string   // 插件二进制存放目录（如 ./plugins/）
	staticDir  string   // 前端 bundle 存放目录（如 ./static/plugins/）
	db         *gorm.DB // 数据库连接
}

// NewInstaller 创建安装器实例
func NewInstaller(pluginsDir, staticDir string, db *gorm.DB) *Installer {
	return &Installer{
		pluginsDir: pluginsDir,
		staticDir:  staticDir,
		db:         db,
	}
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

	// 从 plugin.json 读取插件名
	pjPath := filepath.Join(tempDir, "plugin.json")
	pjData, readErr := os.ReadFile(pjPath)
	if readErr != nil {
		return fmt.Errorf("插件包中缺少 plugin.json 文件")
	}
	var pj pluginJSON
	if jsonErr := json.Unmarshal(pjData, &pj); jsonErr != nil {
		return fmt.Errorf("plugin.json 格式错误: %w", jsonErr)
	}
	if pj.Name == "" {
		return fmt.Errorf("plugin.json 中 name 字段不能为空")
	}
	name := pj.Name

	// 检查是否已安装同名插件
	var existCount int64
	ins.db.Model(&models.SysPlugin{}).Where("name = ?", name).Count(&existCount)
	if existCount > 0 {
		return fmt.Errorf("插件 %s 已安装，请先卸载后再安装", name)
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

	// 处理前端 bundle
	var frontendPath string
	frontendSrc := filepath.Join(destDir, "frontend", "dist")
	if info, statErr := os.Stat(frontendSrc); statErr == nil && info.IsDir() {
		frontendDest := filepath.Join(ins.staticDir, name)
		if cpErr := copyDir(frontendSrc, frontendDest); cpErr != nil {
			return fmt.Errorf("复制前端 bundle 失败: %w", cpErr)
		}
		frontendPath = frontendDest
	}

	// 写入数据库记录
	version := pj.Version
	if version == "" {
		version = "0.0.1"
	}
	now := time.Now()
	record := &models.SysPlugin{
		Name:         name,
		Version:      version,
		Description:  pj.Description,
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

// Uninstall 卸载插件
// cleanData: 是否清除插件相关数据表（{name}_* 表）
func (ins *Installer) Uninstall(ctx context.Context, name string, cleanData bool) error {
	// 尝试 kill 可能残留的插件进程（Windows 上 exe 被占用时无法删除）
	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}
	killPluginProcess(binaryName)

	// 短暂等待进程退出
	time.Sleep(500 * time.Millisecond)

	// 删除插件目录
	pluginDir := filepath.Join(ins.pluginsDir, name)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("删除插件目录失败: %w", err)
	}

	// 删除前端静态文件目录
	staticPluginDir := filepath.Join(ins.staticDir, name)
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

// RegisterMenus 将插件菜单写入 sys_menu 表（挂在"扩展功能"目录下）
func (ins *Installer) RegisterMenus(pluginName string, menus []*proto.MenuItem) error {
	if len(menus) == 0 {
		return nil
	}

	// 确保"扩展功能"父目录存在
	extensionParentId := ins.ensureExtensionMenu()
	if extensionParentId == 0 {
		return fmt.Errorf("创建扩展功能菜单失败")
	}

	// 使用 plugin_{name} 作为唯一标识
	parentMenuName := "plugin_" + pluginName

	// 检查是否已注册
	var existCount int64
	ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", parentMenuName).Count(&existCount)
	if existCount > 0 {
		return nil
	}

	firstMenu := menus[0]

	// 情况 1：只有一个菜单项且无子菜单 → 直接作为叶子菜单
	if len(menus) == 1 && len(firstMenu.Children) == 0 {
		return ins.db.Exec(`INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'C', '无', '', ?, false, '', 'plugin/container', ?, '0', '0', 1, 1, NOW(), NOW())`,
			parentMenuName, firstMenu.Title, firstMenu.Icon, firstMenu.Path,
			fmt.Sprintf("/0/%d/", extensionParentId), extensionParentId, firstMenu.Sort,
		).Error
	}

	// 情况 2：有子菜单 → 创建目录 + 子菜单
	if len(menus) == 1 && len(firstMenu.Children) > 0 {
		// 创建插件目录（M 类型）
		err := ins.db.Exec(`INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'M', '无', '', ?, false, '', '', ?, '0', '0', 1, 1, NOW(), NOW())`,
			parentMenuName, firstMenu.Title, firstMenu.Icon, firstMenu.Path,
			fmt.Sprintf("/0/%d/", extensionParentId), extensionParentId, firstMenu.Sort,
		).Error
		if err != nil {
			return err
		}

		// 获取插件目录 ID
		var pluginDirId int
		ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", parentMenuName).Select("menu_id").Scan(&pluginDirId)
		if pluginDirId == 0 {
			return nil
		}

		// 插入子菜单
		for i, child := range firstMenu.Children {
			childName := fmt.Sprintf("plugin_%s_%d", pluginName, i)
			ins.db.Exec(`INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'C', '无', '', ?, false, '', 'plugin/container', ?, '0', '0', 1, 1, NOW(), NOW())`,
				childName, child.Title, child.Icon, child.Path,
				fmt.Sprintf("/0/%d/%d/", extensionParentId, pluginDirId), pluginDirId, child.Sort,
			)
		}
		return nil
	}

	// 情况 3：多个顶级菜单项 → 创建目录 + 每个作为子菜单
	err := ins.db.Exec(`INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'M', '无', '', ?, false, '', '', ?, '0', '0', 1, 1, NOW(), NOW())`,
		parentMenuName, firstMenu.Title, firstMenu.Icon, "/plugin/"+pluginName,
		fmt.Sprintf("/0/%d/", extensionParentId), extensionParentId, firstMenu.Sort,
	).Error
	if err != nil {
		return err
	}

	var pluginDirId int
	ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", parentMenuName).Select("menu_id").Scan(&pluginDirId)
	if pluginDirId == 0 {
		return nil
	}

	for i, m := range menus {
		childName := fmt.Sprintf("plugin_%s_%d", pluginName, i)
		ins.db.Exec(`INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 'C', '无', '', ?, false, '', 'plugin/container', ?, '0', '0', 1, 1, NOW(), NOW())`,
			childName, m.Title, m.Icon, m.Path,
			fmt.Sprintf("/0/%d/%d/", extensionParentId, pluginDirId), pluginDirId, m.Sort,
		)
	}

	return nil
}

// ensureExtensionMenu 确保"扩展功能"顶级目录存在，返回其 menu_id
func (ins *Installer) ensureExtensionMenu() int {
	const menuName = "PluginExtensions"
	var menuId int
	ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", menuName).Select("menu_id").Scan(&menuId)
	if menuId > 0 {
		return menuId
	}

	// 创建"扩展功能"顶级目录
	ins.db.Exec(`INSERT INTO sys_menu (menu_name, title, icon, path, paths, menu_type, action, permission, parent_id, no_cache, breadcrumb, component, sort, visible, is_frame, create_by, update_by, created_at, updated_at) VALUES (?, '扩展功能', 'ep:grid', '/extensions', '/0/', 'M', '无', '', 0, false, '', 'Layout', 100, '0', '0', 1, 1, NOW(), NOW())`,
		menuName,
	)

	ins.db.Table("sys_menu").Where("menu_name = ? AND deleted_at IS NULL", menuName).Select("menu_id").Scan(&menuId)
	return menuId
}

// UnregisterMenus 从 sys_menu 表中删除插件菜单（软删除）
func (ins *Installer) UnregisterMenus(pluginName string) error {
	prefix := "plugin_" + pluginName + "%"
	return ins.db.Exec("UPDATE sys_menu SET deleted_at = NOW() WHERE menu_name LIKE ? AND deleted_at IS NULL", prefix).Error
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
