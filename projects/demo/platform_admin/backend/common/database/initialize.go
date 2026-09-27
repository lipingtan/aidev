package database

import (
	"fmt"
	"strings"
	"time"

	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	toolsConfig "github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	mycasbin "github.com/go-admin-team/go-admin-core/sdk/pkg/casbin"
	toolsDB "github.com/go-admin-team/go-admin-core/tools/database"
	. "github.com/go-admin-team/go-admin-core/tools/gorm/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"go-admin/common/global"
)

// Setup 配置数据库
func Setup() {
	for k := range toolsConfig.DatabasesConfig {
		setupSimpleDatabase(k, toolsConfig.DatabasesConfig[k])
	}
}

func setupSimpleDatabase(host string, c *toolsConfig.Database) {
	if global.Driver == "" {
		global.Driver = c.Driver
	}
	log.Infof("%s => %s", host, pkg.Green(maskDSN(c.Source)))

	// MySQL: 如果数据库不存在则自动创建
	if c.Driver == "mysql" {
		createDatabaseIfNotExists(c.Source)
	}

	registers := make([]toolsDB.ResolverConfigure, len(c.Registers))
	for i := range c.Registers {
		registers[i] = toolsDB.NewResolverConfigure(
			c.Registers[i].Sources,
			c.Registers[i].Replicas,
			c.Registers[i].Policy,
			c.Registers[i].Tables)
	}
	resolverConfig := toolsDB.NewConfigure(c.Source, c.MaxIdleConns, c.MaxOpenConns, c.ConnMaxIdleTime, c.ConnMaxLifeTime, registers)
	db, err := resolverConfig.Init(&gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: New(
			logger.Config{
				SlowThreshold: time.Second,
				Colorful:      true,
				LogLevel: logger.LogLevel(
					log.DefaultLogger.Options().Level.LevelForGorm()),
			},
		),
	}, opens[c.Driver])

	if err != nil {
		log.Fatal(pkg.Red(c.Driver+" connect error :"), err)
	} else {
		log.Info(pkg.Green(c.Driver + " connect success !"))
	}

	e := mycasbin.Setup(db, "")

	sdk.Runtime.SetDb(host, db)
	sdk.Runtime.SetCasbin(host, e)
}

// maskDSN 将 DSN 中的密码替换为 ***，避免敏感信息打印到日志
// 支持格式：user:password@tcp(host:port)/dbname?params
func maskDSN(dsn string) string {
	// 找到 @ 符号位置
	atIdx := strings.Index(dsn, "@")
	if atIdx < 0 {
		return dsn
	}
	// 找到 : 分隔用户名和密码
	colonIdx := strings.Index(dsn[:atIdx], ":")
	if colonIdx < 0 {
		return dsn
	}
	return dsn[:colonIdx+1] + "***" + dsn[atIdx:]
}

// createDatabaseIfNotExists 检查并创建 MySQL 数据库
func createDatabaseIfNotExists(dsn string) {
	// 解析 DSN 获取主机、用户、密码和数据库名
	// DSN 格式: user:password@tcp(host:port)/dbname?params
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return
	}

	auth := parts[0]
	rest := parts[1]

	// 分离 host:port/dbname?params
	tcpParts := strings.SplitN(rest, ")", 1)
	if len(tcpParts) != 2 {
		return
	}

	hostPart := strings.TrimPrefix(tcpParts[0], "tcp(")
	dbPart := strings.TrimPrefix(tcpParts[1], "/")
	dbName := strings.SplitN(dbPart, "?", 1)[0]

	// 连接到 MySQL 服务器（不指定数据库）
	serverDSN := fmt.Sprintf("%s@tcp(%s)/?charset=utf8mb4", auth, hostPart)
	db, err := gorm.Open(mysql.Open(serverDSN), &gorm.Config{})
	if err != nil {
		log.Warnf("无法连接到 MySQL 服务器以检查数据库: %v", err)
		return
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 检查数据库是否存在
	var count int
	err = db.Raw("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", dbName).Scan(&count).Error
	if err != nil {
		log.Warnf("检查数据库是否存在时出错: %v", err)
		return
	}

	if count == 0 {
		// 创建数据库，使用 utf8mb4 和 utf8mb4_0900_ai_ci 排序规则
		createSQL := fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci",
			dbName,
		)
		err = db.Exec(createSQL).Error
		if err != nil {
			log.Warnf("创建数据库 %s 时出错: %v", dbName, err)
		} else {
			log.Infof("已自动创建数据库: %s (utf8mb4/utf8mb4_0900_ai_ci)", dbName)
		}
	} else {
		log.Infof("数据库已存在: %s", dbName)
	}
}
