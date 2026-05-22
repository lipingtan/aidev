package setup

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gopkg.in/yaml.v3"

	adminModels "go-admin/app/admin/models"
	jobModels "go-admin/app/jobs/models"
	pluginModels "go-admin/app/plugin/models"
	tenantModels "go-admin/app/tenant/models"
	commonModels "go-admin/common/models"
)

// installed 原子标志：0=未安装，1=已安装
var installed int32

// onInstalledCallbacks 安装完成后的回调（用于初始化数据库连接等）
var onInstalledCallbacks []func(dsn string) error

// IsInstalled 检查是否已完成安装
func IsInstalled() bool {
	return atomic.LoadInt32(&installed) == 1
}

// MarkInstalled 标记为已安装
func markInstalled() {
	atomic.StoreInt32(&installed, 1)
}

// RegisterOnInstalled 注册安装完成回调
func RegisterOnInstalled(fn func(dsn string) error) {
	onInstalledCallbacks = append(onInstalledCallbacks, fn)
}

// CheckInstalled 启动时检查是否已安装
func CheckInstalled() {
	if _, err := os.Stat("config/settings.yml"); err == nil {
		markInstalled()
	}
}

// InstallMiddleware 安装检查中间件
// 未安装时拦截所有非 /setup 请求，重定向到安装页面
func InstallMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsInstalled() {
			c.Next()
			return
		}
		// 未安装：只允许 /setup 路径
		if len(c.Request.URL.Path) >= 6 && c.Request.URL.Path[:6] == "/setup" {
			c.Next()
			return
		}
		// 登录接口始终放行（前端适配路由）
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/refresh-token" {
			c.Next()
			return
		}
		// API 请求返回 JSON 错误
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code": 503,
				"msg":  "服务尚未完成初始化，请访问 /setup 完成安装配置",
			})
			c.Abort()
			return
		}
		// 页面请求重定向到安装向导
		c.Redirect(http.StatusFound, "/setup")
		c.Abort()
	}
}

// SetupRequest 安装请求参数
type SetupRequest struct {
	// 数据库配置
	DBHost     string `json:"dbHost" binding:"required"`
	DBPort     int    `json:"dbPort" binding:"required"`
	DBName     string `json:"dbName" binding:"required"`
	DBUser     string `json:"dbUser" binding:"required"`
	DBPassword string `json:"dbPassword"`
	DBCharset  string `json:"dbCharset"`

	// 服务配置
	AppName string `json:"appName"`
	AppPort int    `json:"appPort"`
}

// TestDBRequest 测试数据库连接请求
type TestDBRequest struct {
	DBHost     string `json:"dbHost" binding:"required"`
	DBPort     int    `json:"dbPort" binding:"required"`
	DBName     string `json:"dbName" binding:"required"`
	DBUser     string `json:"dbUser" binding:"required"`
	DBPassword string `json:"dbPassword"`
}

// RegisterRoutes 注册安装向导路由
func RegisterRoutes(r *gin.Engine) {
	setup := r.Group("/setup")
	{
		// 安装向导页面
		setup.GET("", showSetupPage)
		setup.GET("/", showSetupPage)

		// 测试数据库连接
		setup.POST("/test-db", testDBConnection)

		// 执行安装
		setup.POST("/init", doInstall)

		// 检查安装状态
		setup.GET("/status", getStatus)
	}
}

// showSetupPage 返回安装向导 HTML 页面
func showSetupPage(c *gin.Context) {
	if IsInstalled() {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.New("setup").Parse(setupPageHTML))
	_ = tmpl.Execute(c.Writer, nil)
}

// getStatus 获取安装状态
func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"installed": IsInstalled(),
	})
}

// testDBConnection 测试数据库连接（同时验证能否创建数据库）
func testDBConnection(c *gin.Context) {
	var req TestDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 先测试能否连接 MySQL 根节点
	rootDSN := buildRootDSN(req.DBUser, req.DBPassword, req.DBHost, req.DBPort)
	db, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  fmt.Sprintf("无法连接 MySQL 服务器 %s:%d: %s", req.DBHost, req.DBPort, err.Error()),
		})
		return
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err = sqlDB.Ping(); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库连接失败: " + err.Error()})
		return
	}

	// 检查用户是否有创建数据库的权限（尝试 CREATE DATABASE）
	testDB := req.DBName + "_test_permission_check"
	_ = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	_ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))

	// 检查目标数据库是否已存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", req.DBName).Scan(&count)

	msg := fmt.Sprintf("MySQL 连接成功（%s:%d）", req.DBHost, req.DBPort)
	if count > 0 {
		msg += fmt.Sprintf("，数据库 `%s` 已存在", req.DBName)
	} else {
		msg += fmt.Sprintf("，数据库 `%s` 不存在，安装时将自动创建", req.DBName)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": msg})
}

// doInstall 执行安装
func doInstall(c *gin.Context) {
	if IsInstalled() {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "系统已完成安装"})
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	charset := req.DBCharset
	if charset == "" {
		charset = "utf8mb4"
	}
	appName := req.AppName
	if appName == "" {
		appName = "管理平台开发底座"
	}
	appPort := req.AppPort
	if appPort == 0 {
		appPort = 8000
	}

	// 1. 先连接 MySQL（不指定数据库），创建数据库
	if err := createDatabaseIfNotExists(req.DBUser, req.DBPassword, req.DBHost, req.DBPort, req.DBName, charset); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库创建失败: " + err.Error()})
		return
	}

	dsn := buildDSN(req.DBUser, req.DBPassword, req.DBHost, req.DBPort, req.DBName, charset)

	// 2. 连接目标数据库（禁用外键约束，避免 AutoMigrate 建外键时类型不兼容）
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库连接失败: " + err.Error()})
		return
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 3. 执行数据库迁移（建表）
	migrateDb := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	if err = runMigrations(migrateDb); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库初始化失败: " + err.Error()})
		return
	}

	// 4. 检查是否已有数据（sys_user 表有记录则跳过初始数据写入）
	var userCount int64
	db.Raw("SELECT COUNT(*) FROM sys_user").Scan(&userCount)
	if userCount == 0 {
		// 写入初始数据（角色、菜单、管理员账号等）
		if err = adminModels.InitDb(db); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "初始数据写入失败: " + err.Error()})
			return
		}
	}

	// 4. 写入配置文件
	if err = writeConfig(req, dsn); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "配置文件写入失败: " + err.Error()})
		return
	}

	// 5. 执行安装完成回调（初始化主数据库连接、Casbin 等）
	for _, fn := range onInstalledCallbacks {
		if err = fn(dsn); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "服务初始化失败: " + err.Error()})
			return
		}
	}

	// 6. 标记为已安装
	markInstalled()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "安装成功！系统已就绪，正在跳转...",
		"data": gin.H{"redirect": "/"},
	})
}

// runMigrations 执行数据库迁移
func runMigrations(db *gorm.DB) error {
	// go-admin 系统表
	if err := db.AutoMigrate(
		&adminModels.SysUser{},
		&adminModels.SysRole{},
		&adminModels.SysMenu{},
		&adminModels.SysDept{},
		&adminModels.SysPost{},
		&adminModels.SysConfig{},
		&adminModels.SysApi{},
		&adminModels.SysDictType{},
		&adminModels.SysDictData{},
		&adminModels.SysLoginLog{},
		&adminModels.SysOperaLog{},
		&commonModels.Migration{},
		&jobModels.SysJob{},
	); err != nil {
		return fmt.Errorf("系统表迁移失败: %w", err)
	}

	// 租户表
	if err := db.AutoMigrate(
		&tenantModels.Tenant{},
	); err != nil {
		return fmt.Errorf("租户表迁移失败: %w", err)
	}

	// 插件管理表
	if err := db.AutoMigrate(
		&pluginModels.SysPlugin{},
	); err != nil {
		return fmt.Errorf("插件管理表迁移失败: %w", err)
	}

	return nil
}

// writeConfig 写入 settings.yml 配置文件
func writeConfig(req SetupRequest, _ string) error {
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	cfg := map[string]interface{}{
		"settings": map[string]interface{}{
			"application": map[string]interface{}{
				"name":          req.AppName,
				"mode":          "prod",
				"host":          "0.0.0.0",
				"port":          req.AppPort,
				"readtimeout":   1,
				"writertimeout": 2,
			},
			"logger": map[string]interface{}{
				"path":  "temp/logs",
				"level": "info",
			},
			"jwt": map[string]interface{}{
				"secret":  generateSecret(),
				"timeout": 86400,  // 24小时
				"refresh": 604800, // 7天
			},
			"database": map[string]interface{}{
				"driver":   "mysql",
				"source":   buildDSN(req.DBUser, req.DBPassword, req.DBHost, req.DBPort, req.DBName, "utf8mb4"),
				"register": []string{"default"},
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile("config/settings.yml", data, 0644)
}

// buildDSN 构建 MySQL DSN
func buildDSN(user, password, host string, port int, dbName, charset string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s&sql_mode=''",
		user, password, host, port, dbName, charset)
}

// buildRootDSN 构建不指定数据库的 MySQL DSN（用于创建数据库）
func buildRootDSN(user, password, host string, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=True&loc=Local&timeout=10s",
		user, password, host, port)
}

// createDatabaseIfNotExists 连接 MySQL 根节点，创建数据库（如果不存在）
func createDatabaseIfNotExists(user, password, host string, port int, dbName, charset string) error {
	// 连接 MySQL，不指定数据库
	rootDSN := buildRootDSN(user, password, host, port)
	db, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("无法连接 MySQL 服务器 %s:%d，请检查地址、端口和账号密码: %w", host, port, err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 创建数据库
	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET %s DEFAULT COLLATE %s_unicode_ci",
		dbName, charset, charset,
	)
	if err = db.Exec(createSQL).Error; err != nil {
		return fmt.Errorf("创建数据库 `%s` 失败: %w", dbName, err)
	}

	return nil
}

// generateSecret 生成随机 JWT 密钥
func generateSecret() string {
	return fmt.Sprintf("game-server-%d", time.Now().UnixNano())
}
