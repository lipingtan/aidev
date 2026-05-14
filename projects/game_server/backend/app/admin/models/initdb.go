package models

import (
	_ "embed"
	"log"
	"strings"

	"gorm.io/gorm"
)

//go:embed sql/db.sql
var dbSQL string

// InitDb 执行初始数据 SQL（从 embed 读取，不依赖磁盘文件）
func InitDb(db *gorm.DB) error {
	return execSQLContent(db, dbSQL)
}

// execSQLContent 执行 SQL 内容
func execSQLContent(db *gorm.DB, content string) error {
	// 去掉 UTF-8 BOM
	content = strings.TrimPrefix(content, "\xef\xbb\xbf")

	// 关闭严格模式，兼容旧版 db.sql 的列顺序差异
	_ = db.Exec("SET sql_mode=''").Error

	// 按行过滤注释，再重新拼接
	var filteredLines []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	content = strings.Join(filteredLines, "\n")

	sqlList := strings.Split(content, ";")
	for _, s := range sqlList {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if err := db.Exec(s).Error; err != nil {
			log.Printf("SQL 执行警告（已跳过）: %v | sql: %.120s", err, s)
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "Query was empty") {
				continue
			}
			return err
		}
	}
	return nil
}
