package storage

import "io"

// FileStore 文件存储接口，支持本地磁盘和 OSS 等多种实现
type FileStore interface {
	// Save 保存文件到指定分类目录
	Save(category string, filename string, reader io.Reader) (path string, err error)
	// GetURL 获取文件访问 URL（本地返回相对路径，OSS 返回签名 URL）
	GetURL(path string) (url string, err error)
	// Delete 删除文件
	Delete(path string) error
	// Exists 检查文件是否存在
	Exists(path string) bool
}
