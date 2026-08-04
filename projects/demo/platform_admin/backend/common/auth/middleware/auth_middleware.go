package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-admin/common/auth/service"
	"go-admin/common/auth/strategy"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk"
)

const authContextKey = "auth_context"

// BizUserChecker C端用户状态检查接口（由 app/user_auth 注入实现，避免 common → app 反向依赖）
type BizUserChecker interface {
	// CheckBizUser 查询C端用户的状态和 token 版本号
	// 返回 status（1=正常）、tokenVersion、error
	CheckBizUser(userID int64) (status int, tokenVersion int, err error)
}

// bizUserCacheEntry 缓存条目
type bizUserCacheEntry struct {
	Status       int
	TokenVersion int
	CachedAt     time.Time
}

// bizUserCache C端用户信息缓存（TTL 5 分钟，进程内）
var (
	bizUserCacheMap sync.Map
	bizUserCacheTTL = 5 * time.Minute
)

// bizUserInvalidateKey 生成 Redis 失效标记 key
func bizUserInvalidateKey(userID int64) string {
	return fmt.Sprintf("biz_user:invalidate:%d", userID)
}

// InvalidateBizUserCache 失效指定用户的本地缓存，同时写 Redis 失效标记（跨 Pod 有效）
// TTL=5分钟与本地缓存 TTL 一致，确保其他 Pod 在此期间不命中本地缓存
func InvalidateBizUserCache(userID int64) {
	// 清本进程内存缓存
	bizUserCacheMap.Delete(userID)
	// 写 Redis 失效标记（跨 Pod 通知）
	if adapter := sdk.Runtime.GetCacheAdapter(); adapter != nil {
		_ = adapter.Set(bizUserInvalidateKey(userID), "1", int(bizUserCacheTTL.Seconds()))
	}
}

// getBizUserCached 带缓存查询C端用户信息
// 有 Redis 失效标记时强制穿透查库，保证跨 Pod 强制下线的实时性
func getBizUserCached(checker BizUserChecker, userID int64) (status int, tokenVersion int, err error) {
	// 检查 Redis 失效标记（跨 Pod 强制下线）
	if adapter := sdk.Runtime.GetCacheAdapter(); adapter != nil {
		if val, e := adapter.Get(bizUserInvalidateKey(userID)); e == nil && val != "" {
			// 有失效标记：清本地缓存，穿透查库
			bizUserCacheMap.Delete(userID)
			goto queryDB
		}
	}

	// 尝试从进程内缓存获取
	if val, ok := bizUserCacheMap.Load(userID); ok {
		entry := val.(*bizUserCacheEntry)
		if time.Since(entry.CachedAt) < bizUserCacheTTL {
			return entry.Status, entry.TokenVersion, nil
		}
		// 过期，删除缓存
		bizUserCacheMap.Delete(userID)
	}

queryDB:
	// 缓存 miss 或失效，查库
	status, tokenVersion, err = checker.CheckBizUser(userID)
	if err != nil {
		return 0, 0, err
	}

	// 写入进程内缓存
	bizUserCacheMap.Store(userID, &bizUserCacheEntry{
		Status:       status,
		TokenVersion: tokenVersion,
		CachedAt:     time.Now(),
	})
	return status, tokenVersion, nil
}

// AuthContext 认证上下文，注入到 gin.Context
type AuthContext struct {
	UserID       int64
	TenantID     int64
	Roles        []int64
	UserPool     string // "admin" | "user"，旧 token 无此字段时为 "admin"
	TokenVersion int    // token 版本号，旧 token 无此字段时为 0
}

// SetAuthContext 设置认证上下文到 gin.Context
func SetAuthContext(c *gin.Context, ctx *AuthContext) {
	c.Set(authContextKey, ctx)
}

// GetAuthContext 从 gin.Context 获取认证上下文
func GetAuthContext(c *gin.Context) *AuthContext {
	val, exists := c.Get(authContextKey)
	if !exists {
		return nil
	}
	ctx, ok := val.(*AuthContext)
	if !ok {
		return nil
	}
	return ctx
}

// AuthMiddleware 认证中间件
// 解析 access_token、检查黑名单、注入 AuthContext
// bizUserChecker 可选，传 nil 则跳过C端用户额外校验（兼容未启用C端认证的部署）
func AuthMiddleware(authSvc *service.AuthService, bizUserChecker ...BizUserChecker) gin.HandlerFunc {
	var checker BizUserChecker
	if len(bizUserChecker) > 0 && bizUserChecker[0] != nil {
		checker = bizUserChecker[0]
	}
	return func(c *gin.Context) {
		// 从 Authorization header 获取 Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"data":    nil,
				"message": "缺少认证信息",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"data":    nil,
				"message": "认证格式错误",
			})
			return
		}

		tokenStr := parts[1]

		// 解析 access_token（使用 strategy 包的公共函数）
		claims, err := strategy.ParseAccessToken(authSvc.GetConfig(), tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40102,
				"data":    nil,
				"message": "token 无效或已过期",
			})
			return
		}

		// 检查黑名单
		if authSvc.IsBlacklisted(claims.ID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40103,
				"data":    nil,
				"message": "token 已失效",
			})
			return
		}

		// 检查租户状态（四态：1=ACTIVE, 2=READ_ONLY, 0=DISABLED, 3=CANCELLING）
		tenantStatus, err := authSvc.CheckTenantStatus(claims.TenantID)
		if err == nil {
			// 自动降级：status=1 且 expired_at 已过期 → CAS 更新为 READ_ONLY
			if tenantStatus == 1 {
				if degraded := authSvc.AutoDegradeIfExpired(claims.TenantID); degraded {
					tenantStatus = 2
				}
			}

			switch tenantStatus {
			case 1: // ACTIVE → 放行
			case 2: // READ_ONLY → 仅 GET/HEAD 放行
				if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"code":    40305,
						"data":    nil,
						"message": "租户已降为只读模式",
					})
					return
				}
			case 0, 3: // DISABLED, CANCELLING → 全部拒绝
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"code":    40105,
					"data":    nil,
					"message": "租户已禁用",
				})
				return
			}
		}

		// 检查用户状态
		userStatus, err := authSvc.CheckUserStatus(claims.UserID)
		if err == nil && userStatus != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    40104,
				"data":    nil,
				"message": "账号已禁用",
			})
			return
		}

		// 兼容处理：UserPool="" 视为 "admin"（旧 token 无此字段）
		userPool := claims.UserPool
		if userPool == "" {
			userPool = strategy.UserPoolAdmin
		}

		// 注入 AuthContext
		authCtx := &AuthContext{
			UserID:       claims.UserID,
			TenantID:     claims.TenantID,
			Roles:        claims.Roles,
			UserPool:     userPool,
			TokenVersion: claims.TokenVersion,
		}
		SetAuthContext(c, authCtx)

		// C端用户额外校验：token_version + status
		if userPool == strategy.UserPoolUser && checker != nil {
			status, dbTokenVersion, err := getBizUserCached(checker, claims.UserID)
			if err != nil {
				// DB 查询失败 → fail-closed（拒绝）
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    40101,
					"data":    nil,
					"message": "用户信息查询失败",
				})
				return
			}
			if status != 1 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    40104,
					"data":    nil,
					"message": "账号已禁用",
				})
				return
			}
			if dbTokenVersion != claims.TokenVersion {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    40103,
					"data":    nil,
					"message": "token 已失效",
				})
				return
			}
		}

		c.Next()
	}
}
