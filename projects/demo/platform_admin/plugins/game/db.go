package main

import (
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

// initDB 初始化数据库连接（只执行一次），并自动迁移所有表结构
func initDB(dsn string) error {
	var initErr error
	dbOnce.Do(func() {
		var err error
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			initErr = err
			return
		}
		// 自动迁移所有模型表
		initErr = db.AutoMigrate(
			&Game{},
			&Dlc{},
			&Player{},
			&Save{},
			&Order{},
			&PaymentConfig{},
			&H5Page{},
		)
	})
	return initErr
}
