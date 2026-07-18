package storage

import (
	"io"
	"os"
	"path/filepath"
)

// LocalStore 本地磁盘文件存储实现
type LocalStore struct {
	BaseDir string // 存储根目录，默认 "static/uploadfile"
}

// NewLocalStore 创建本地存储实例
func NewLocalStore(baseDir string) *LocalStore {
	if baseDir == "" {
		baseDir = "static/uploadfile"
	}
	return &LocalStore{BaseDir: baseDir}
}

// Save 将文件存储到 BaseDir/category/filename，自动创建目录
func (s *LocalStore) Save(category string, filename string, reader io.Reader) (string, error) {
	dir := filepath.Join(s.BaseDir, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(dir, filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", err
	}

	// 返回相对路径：category/filename
	return filepath.ToSlash(filepath.Join(category, filename)), nil
}

// GetURL 返回静态文件访问路径，供前端直接访问
func (s *LocalStore) GetURL(path string) (string, error) {
	return "/static/uploadfile/" + path, nil
}

// Delete 删除指定路径的文件
func (s *LocalStore) Delete(path string) error {
	fullPath := filepath.Join(s.BaseDir, path)
	return os.Remove(fullPath)
}

// Exists 检查文件是否存在
func (s *LocalStore) Exists(path string) bool {
	fullPath := filepath.Join(s.BaseDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}
