package api

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/config/source/file"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gorm.io/gorm"

	"go-admin/app/admin/models"
	"go-admin/app/admin/router"
	"go-admin/app/jobs"
	"go-admin/app/setup"
	authPkg "go-admin/common/auth"
	authConfig "go-admin/common/auth/config"
	"go-admin/common/database"
	"go-admin/common/global"
	common "go-admin/common/middleware"
	"go-admin/common/middleware/handler"
	"go-admin/common/storage"
	ext "go-admin/config"
	"go-admin/web"
)

var (
	configYml string
	apiCheck  bool
	StartCmd  = &cobra.Command{
		Use:          "server",
		Short:        "Start API server",
		Example:      "go-admin server -c config/settings.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			preRun()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

var AppRouters = make([]func(), 0)

func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/settings.yml", "Start server with provided configuration file")
	StartCmd.PersistentFlags().BoolVarP(&apiCheck, "api", "a", false, "Start server with check api data")
	AppRouters = append(AppRouters, router.InitRouter)
}

func preRun() {
	// 检查是否已安装
	setup.CheckInstalled()

	// 已安装才加载配置和初始化数据库
	if setup.IsInstalled() {
		config.ExtendConfig = &ext.ExtConfig
		config.Setup(
			file.NewSource(file.WithPath(configYml)),
			database.Setup,
			storage.Setup,
		)
		queue := sdk.Runtime.GetMemoryQueue("")
		queue.Register(global.LoginLog, models.SaveLoginLog)
		queue.Register(global.OperateLog, models.SaveOperaLog)
		queue.Register(global.ApiCheck, models.SaveSysApi)
		go queue.Run()
	}

	log.Info("starting api server...")
}

func run() error {
	if config.ApplicationConfig != nil && config.ApplicationConfig.Mode == pkg.ModeProd.String() {
		gin.SetMode(gin.ReleaseMode)
	}
	initRouter()

	// 注册安装向导路由（始终注册）
	if r, ok := sdk.Runtime.GetEngine().(*gin.Engine); ok {
		setup.RegisterRoutes(r)
	}

	// 已安装才注册业务路由
	if setup.IsInstalled() {
		for _, f := range AppRouters {
			f()
		}

		// 初始化 auth-rbac 模块（注册新认证路由和中间件）
		if r, ok := sdk.Runtime.GetEngine().(*gin.Engine); ok {
			var db *gorm.DB
			for _, d := range sdk.Runtime.GetDb() {
				if d != nil {
					db = d
					break
				}
			}
			if db != nil {
				authCfg := authConfig.DefaultConfig()
				if config.JwtConfig != nil && config.JwtConfig.Secret != "" {
					authCfg.JWT.Secret = config.JwtConfig.Secret
				}
				if err := authPkg.Init(authCfg, db, r); err != nil {
					log.Fatalf("auth-rbac 初始化失败: %v", err)
				}
			}
		}
	} else {
		log.Info("系统未安装，请访问 /setup 完成初始化")
	}

	// 安装完成回调：动态初始化数据库和业务路由（无需重启）
	setup.RegisterOnInstalled(func(dsn string) error {
		config.ExtendConfig = &ext.ExtConfig
		config.Setup(
			file.NewSource(file.WithPath(configYml)),
			database.Setup,
			storage.Setup,
		)
		queue := sdk.Runtime.GetMemoryQueue("")
		queue.Register(global.LoginLog, models.SaveLoginLog)
		queue.Register(global.OperateLog, models.SaveOperaLog)
		queue.Register(global.ApiCheck, models.SaveSysApi)
		go queue.Run()

		for _, f := range AppRouters {
			f()
		}

		// 初始化 auth-rbac 模块（注册新认证路由和中间件）
		if r, ok := sdk.Runtime.GetEngine().(*gin.Engine); ok {
			var db *gorm.DB
			for _, d := range sdk.Runtime.GetDb() {
				if d != nil {
					db = d
					break
				}
			}
			if db != nil {
				authCfg := authConfig.DefaultConfig()
				if config.JwtConfig != nil && config.JwtConfig.Secret != "" {
					authCfg.JWT.Secret = config.JwtConfig.Secret
				}
				if err := authPkg.Init(authCfg, db, r); err != nil {
					log.Fatalf("auth-rbac 初始化失败: %v", err)
				}
			}
		}

		go func() {
			jobs.InitJob()
			jobs.Setup(sdk.Runtime.GetDb())
		}()

		log.Info("安装完成，业务路由已动态注册")
		return nil
	})

	// 未安装时使用默认端口 8000
	host := "0.0.0.0"
	port := 8000
	if setup.IsInstalled() && config.ApplicationConfig != nil {
		host = config.ApplicationConfig.Host
		port = int(config.ApplicationConfig.Port)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: sdk.Runtime.GetEngine(),
	}
	if setup.IsInstalled() && config.ApplicationConfig != nil {
		srv.ReadTimeout  = time.Duration(config.ApplicationConfig.ReadTimeout) * time.Second
		srv.WriteTimeout = time.Duration(config.ApplicationConfig.WriterTimeout) * time.Second
	}

	if setup.IsInstalled() {
		go func() {
			jobs.InitJob()
			jobs.Setup(sdk.Runtime.GetDb())
		}()
	}

	if apiCheck && setup.IsInstalled() {
		var routers = sdk.Runtime.GetRouter()
		q := sdk.Runtime.GetMemoryQueue("")
		mp := make(map[string]interface{})
		mp["List"] = routers
		message, err := sdk.Runtime.GetStreamMessage("", global.ApiCheck, mp)
		if err != nil {
			log.Infof("GetStreamMessage error, %s \n", err.Error())
		} else {
			err = q.Append(message)
			if err != nil {
				log.Infof("Append message error, %s \n", err.Error())
			}
		}
	}

	go func() {
		if config.SslConfig != nil && config.SslConfig.Enable {
			if err := srv.ListenAndServeTLS(config.SslConfig.Pem, config.SslConfig.KeyStr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("listen: ", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("listen: ", err)
			}
		}
	}()

	fmt.Println(pkg.Green("Server run at:"))
	fmt.Printf("-  Local:   http://localhost:%d/ \r\n", port)
	fmt.Printf("-  Network: http://%s:%d/ \r\n", pkg.GetLocalHost(), port)
	if setup.IsInstalled() {
		fmt.Println(pkg.Green("Swagger run at:"))
		fmt.Printf("-  Local:   http://localhost:%d/swagger/admin/index.html \r\n", port)
	} else {
		fmt.Printf("-  安装向导: http://localhost:%d/setup \r\n", port)
	}
	fmt.Printf("%s Enter Control + C Shutdown Server \r\n", pkg.GetCurrentTimeStr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info("Shutdown Server ... ")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Info("Server exiting")
	return nil
}

func tip() {
	usageStr := `欢迎使用 ` + pkg.Green(`管理平台开发底座 `+global.Version) + ` 可以使用 ` + pkg.Red(`-h`) + ` 查看命令`
	fmt.Printf("%s \n\n", usageStr)
}

func initRouter() {
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		h = gin.New()
		sdk.Runtime.SetEngine(h)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
	}
	if config.SslConfig != nil && config.SslConfig.Enable {
		r.Use(handler.TlsHandler())
	}
	r.Use(common.Sentinel()).
		Use(common.RequestId(pkg.TrafficKey)).
		Use(api.SetRequestLogger)

	common.InitMiddleware(r)

	// 安装检查中间件（未安装时拦截业务请求）
	r.Use(setup.InstallMiddleware())

	// 注册前端静态文件服务（SPA 路由支持）
	registerStaticFiles(r)
}

// registerStaticFiles 将嵌入的前端文件注册为静态资源
// SPA 路由：所有非 /api、/setup、/swagger 的请求都返回 index.html 或对应静态文件
func registerStaticFiles(r *gin.Engine) {
	distFS, err := fs.Sub(web.Files, "dist")
	if err != nil {
		log.Warnf("前端文件未嵌入，跳过静态文件服务: %v", err)
		return
	}

	fileServer := http.FileServer(http.FS(distFS))

	// pure-admin Vite 构建产物 + 插件前端 bundle
	// /static/plugins/* 从磁盘读取，其他从嵌入 FS 读取
	r.GET("/static/*filepath", func(c *gin.Context) {
		fp := c.Param("filepath")
		// 插件前端 bundle 从磁盘读取
		if strings.HasPrefix(fp, "/plugins/") {
			filePath := "./static" + fp
			c.Header("Content-Type", "application/javascript")
			c.File(filePath)
			return
		}
		c.Request.URL.Path = "/static" + fp
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Request.URL.Path = "/favicon.ico"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/logo.svg", func(c *gin.Context) {
		c.Request.URL.Path = "/logo.svg"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/platform-config.json", func(c *gin.Context) {
		c.Request.URL.Path = "/platform-config.json"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// SPA 入口：所有非 API 路径返回 index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "接口不存在"})
			return
		}
		if len(path) >= 6 && path[:6] == "/setup" {
			c.Next()
			return
		}
		if len(path) >= 8 && path[:8] == "/swagger" {
			c.Next()
			return
		}
		indexFile, err := web.Files.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "前端文件未找到，请先构建前端")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexFile)
	})
}

