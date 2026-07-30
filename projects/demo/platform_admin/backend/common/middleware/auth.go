package middleware

import (
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/config"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

// AuthInit jwt验证（旧 go-admin 框架，V2 已用 auth-rbac AuthMiddleware 替代）
// 保留函数签名兼容 app/admin/router，待 sys_* 路由清理后删除
func AuthInit() (*jwt.GinJWTMiddleware, error) {
	timeout := time.Hour
	if config.ApplicationConfig != nil && config.ApplicationConfig.Mode == "dev" {
		timeout = time.Duration(876010) * time.Hour
	} else {
		if config.JwtConfig != nil && config.JwtConfig.Timeout != 0 {
			timeout = time.Duration(config.JwtConfig.Timeout) * time.Second
		}
	}
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte(config.JwtConfig.Secret),
		Timeout:     timeout,
		MaxRefresh:  time.Hour,
		TokenLookup: "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:    time.Now,
	})
}