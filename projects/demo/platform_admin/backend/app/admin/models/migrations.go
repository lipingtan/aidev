package models

import "gorm.io/gorm"

// MigrateApprovalTables 注册审批流相关表到 AutoMigrate（CR-7）
func MigrateApprovalTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&AdminApprovalFlow{},
		&AdminApprovalNode{},
		&AdminApproval{},
		&AdminApprovalVote{},
	)
}
