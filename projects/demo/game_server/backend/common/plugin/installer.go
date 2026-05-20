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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go-admin/app/plugin/models"
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
// name: 插件名称
// file: 文件内容读取器
// filename: 原始文件名，用于判断压缩格式（.zip 或 .tar.gz）
func (ins *Installer) InstallFromFile(ctx context.Context, name string, file io.Reader, filename string) error {
	destDir := filepath.Join(ins.pluginsDir, name)

	// 创建插件目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}

	// 根据文件后缀选择解压方式
	var err error
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		err = extractTarGz(file, destDir)
	} else if strings.HasSuffix(filename, ".zip") {
		err = extractZip(file, destDir)
	} else {
		return fmt.Errorf("不支持的文件格式: %s（仅支持 .zip 和 .tar.gz）", filename)
	}
	if err != nil {
		// 解压失败时清理目录
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("解压文件失败: %w", err)
	}

	// 查找二进制文件
	binaryName := name
	if runtime.GOOS == "windows" {
		binaryName = name + ".exe"
	}
	binaryPath := filepath.Join(destDir, binaryName)
	if _, statErr := os.Stat(binaryPath); statErr != nil {
		// 二进制文件不存在不阻断安装，记录相对路径即可
		binaryPath = destDir
	}

	// 处理前端 bundle：如果存在 frontend/dist/ 目录，复制到 staticDir/{name}/
	var frontendPath string
	frontendSrc := filepath.Join(destDir, "frontend", "dist")
	if info, statErr := os.Stat(frontendSrc); statErr == nil && info.IsDir() {
		frontendDest := filepath.Join(ins.staticDir, name)
		if cpErr := copyDir(frontendSrc, frontendDest); cpErr != nil {
			return fmt.Errorf("复制前端 bundle 失败: %w", cpErr)
		}
		frontendPath = frontendDest
	}

	// 读取 plugin.json 获取版本和描述信息
	version := "0.0.1"
	description := ""
	pjPath := filepath.Join(destDir, "plugin.json")
	if data, readErr := os.ReadFile(pjPath); readErr == nil {
		var pj pluginJSON
		if jsonErr := json.Unmarshal(data, &pj); jsonErr == nil {
			if pj.Version != "" {
				version = pj.Version
			}
			if pj.Description != "" {
				description = pj.Description
			}
		}
	}

	// 写入数据库记录
	now := time.Now()
	record := &models.SysPlugin{
		Name:         name,
		Version:      version,
		Description:  description,
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
