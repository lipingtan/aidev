package models

import (
	_ "embed"
	"log"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// MigrateApprovalTables 注册审批流相关表到 AutoMigrate
func MigrateApprovalTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&AdminApprovalFlow{},
		&AdminApproval{},
		&AdminApprovalNode{},
		&AdminApprovalVote{},
	)
}

//go:embed sql/db.sql
var dbSQL string

//go:embed sql/menu_init.sql
var menuInitSQL string

// InitDb 执行初始数据 SQL（从 embed 读取，不依赖磁盘文件）
func InitDb(db *gorm.DB) error {
	if err := execSQLContent(db, dbSQL); err != nil {
		return err
	}
	// 执行游戏管理菜单初始化
	return execSQLContentRaw(db, menuInitSQL)
}

// sysMenuInsertRe 匹配 INSERT INTO sys_menu VALUES 语句（不区分大小写）
var sysMenuInsertRe = regexp.MustCompile(`(?i)INSERT\s+INTO\s+sys_menu\s+VALUES\s*`)

// sysMenuCols sys_menu 表列名（与 db.sql 中 VALUES 的列顺序一致）
const sysMenuCols = "(menu_id,menu_name,title,icon,path,paths,menu_type,action,permission,parent_id,no_cache,breadcrumb,component,sort,visible,is_frame,create_by,update_by,created_at,updated_at,deleted_at)"

// sysUserInsertRe 匹配 INSERT INTO sys_user VALUES 语句
var sysUserInsertRe = regexp.MustCompile(`(?i)INSERT\s+INTO\s+sys_user\s+VALUES\s*`)

// sysUserCols sys_user 表列名（与 db.sql 中 VALUES 的列顺序一致，不含 tenant_id）
const sysUserCols = "(user_id,username,password,nick_name,phone,role_id,salt,avatar,sex,email,dept_id,post_id,remark,status,create_by,update_by,created_at,updated_at,deleted_at)"

// tableColsMap 各表的列名映射（key=表名，value=列名列表，顺序与 db.sql 中 VALUES 一致）
var tableColsMap = map[string]string{
	"sys_api":       "(id,handle,title,path,type,action,created_at,updated_at,deleted_at,create_by,update_by)",
	"sys_config":    "(id,config_name,config_key,config_value,config_type,is_frontend,remark,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_dept":      "(dept_id,parent_id,dept_path,dept_name,sort,leader,phone,email,status,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_dict_data": "(dict_code,dict_sort,dict_label,dict_value,dict_type,css_class,list_class,is_default,status,`default`,remark,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_dict_type": "(dict_id,dict_name,dict_type,status,remark,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_post":      "(post_id,post_name,post_code,sort,status,remark,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_role":      "(role_id,role_name,status,role_key,role_sort,flag,remark,admin,data_scope,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_user":      "(user_id,username,password,nick_name,phone,role_id,salt,avatar,sex,email,dept_id,post_id,remark,status,create_by,update_by,created_at,updated_at,deleted_at)",
	"sys_menu":      "(menu_id,menu_name,title,icon,path,paths,menu_type,action,permission,parent_id,no_cache,breadcrumb,component,sort,visible,is_frame,create_by,update_by,created_at,updated_at,deleted_at)",
}

// tableInsertRe 为每个表生成正则
var tableInsertRe = func() map[string]*regexp.Regexp {
	m := make(map[string]*regexp.Regexp)
	for table := range tableColsMap {
		m[table] = regexp.MustCompile(`(?i)INSERT\s+INTO\s+` + table + `\s+VALUES\s*`)
	}
	return m
}()

// execSQLContent 执行 SQL 内容
func execSQLContent(db *gorm.DB, content string) error {
	// 去掉 UTF-8 BOM
	content = strings.TrimPrefix(content, "\xef\xbb\xbf")

	// 关闭严格模式，兼容旧版 db.sql 的列顺序差异
	_ = db.Exec("SET sql_mode=''").Error

	// 将各表的无列名 INSERT 替换为带列名的形式，避免列顺序不匹配
	for table, cols := range tableColsMap {
		re := tableInsertRe[table]
		content = re.ReplaceAllString(content, "INSERT INTO "+table+" "+cols+" VALUES ")
	}

	// 替换旧版图标为 ep: 格式
	iconReplacements := map[string]string{
		"'api-server'":    "'ep:setting'",
		"'user'":          "'ep:user'",
		"'tree-table'":    "'ep:menu'",
		"'peoples'":       "'ep:avatar'",
		"'tree'":          "'ep:office-building'",
		"'pass'":          "'ep:postcard'",
		"'education'":     "'ep:notebook'",
		"'dev-tools'":     "'ep:tools'",
		"'guide'":         "'ep:document'",
		"'swagger'":       "'ep:operation'",
		"'log'":           "'ep:document-copy'",
		"'logininfor'":    "'ep:tickets'",
		"'skill'":         "'ep:edit'",
		"'druid'":         "'ep:monitor'",
		"'time-range'":    "'ep:timer'",
		"'job'":           "'ep:calendar'",
		"'bug'":           "'ep:document'",
		"'code'":          "'ep:code'",
		"'build'":         "'ep:edit-pen'",
		"'api-doc'":       "'ep:connection'",
		"'system-tools'":  "'ep:set-up'",
		"'upload'":        "''",
		"'app-group-fill'": "''",
	}
	for old, new_ := range iconReplacements {
		content = strings.ReplaceAll(content, old, new_)
	}

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
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "Query was empty") {
				continue
			}
			// 打印完整 SQL 便于排查
			log.Printf("SQL 执行失败: %v\n完整SQL: %s", err, s)
			return err
		}
	}
	return nil
}

// execSQLContentRaw 执行 SQL 内容（不做列名替换，适用于已带列名的 SQL）
func execSQLContentRaw(db *gorm.DB, content string) error {
	content = strings.TrimPrefix(content, "\xef\xbb\xbf")

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
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "Query was empty") {
				continue
			}
			log.Printf("菜单初始化SQL执行失败: %v\nSQL: %s", err, s)
			return err
		}
	}
	return nil
}
