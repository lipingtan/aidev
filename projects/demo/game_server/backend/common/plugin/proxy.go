package plugin

import (
	"io"
	"net/http"
	"strings"

	"game-server/plugin-sdk/proto"

	"github.com/gin-gonic/gin"
)

// PluginProxy HTTP 请求代理，将匹配插件路由的请求转发给对应插件处理
//
// 集成方式：在主路由中调用 RegisterRoutes 注册代理路由
//
//	mgr := plugin.NewPluginManager()
//	proxy := plugin.NewPluginProxy(mgr, dbDsn)
//	apiV1 := r.Group("/api/v1")
//	proxy.RegisterRoutes(apiV1)
type PluginProxy struct {
	mgr   *PluginManager
	dbDsn string // 数据库连接串，注入到插件请求中
}

// NewPluginProxy 创建插件代理实例
func NewPluginProxy(mgr *PluginManager, dbDsn string) *PluginProxy {
	return &PluginProxy{mgr: mgr, dbDsn: dbDsn}
}

// RegisterRoutes 注册插件代理路由到指定路由组
func (p *PluginProxy) RegisterRoutes(r *gin.RouterGroup) {
	r.Any("/plugin/:name/*action", p.Handler())
}

// Handler 返回处理插件代理请求的 gin.HandlerFunc
func (p *PluginProxy) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 URL 中提取插件名和 action 路径
		name := c.Param("name")
		action := c.Param("action")

		// 获取插件实例
		inst, exists := p.mgr.GetPlugin(name)
		if !exists || inst.Status != StatusRunning {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code": 503,
				"msg":  "插件未运行或不存在: " + name,
			})
			c.Abort()
			return
		}

		// 读取请求体
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
		}

		// 提取请求头
		headers := make(map[string]string, len(c.Request.Header))
		for k, v := range c.Request.Header {
			headers[k] = strings.Join(v, ",")
		}

		// 构造 HttpRequest，注入上下文信息
		req := &proto.HttpRequest{
			Method:      c.Request.Method,
			Path:        action,
			Headers:     headers,
			Body:        body,
			Query:       c.Request.URL.RawQuery,
			UserId:      getUserIDFromContext(c),
			TenantId:    getTenantIDFromContext(c),
			Roles:       getRolesFromContext(c),
			Permissions: getPermissionsFromContext(c),
			DbDsn:       p.dbDsn,
		}

		// 调用插件处理请求
		resp, err := inst.Service.HandleRequest(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"code": 502,
				"msg":  "插件处理请求失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将插件响应写回 HTTP
		for k, v := range resp.Headers {
			c.Header(k, v)
		}
		c.Data(int(resp.StatusCode), "", resp.Body)
	}
}

// getUserIDFromContext 从 gin context 中提取用户 ID
func getUserIDFromContext(c *gin.Context) int64 {
	// 尝试 "userId" key
	if v, exists := c.Get("userId"); exists {
		return toInt64(v)
	}
	// 尝试 "user_id" key
	if v, exists := c.Get("user_id"); exists {
		return toInt64(v)
	}
	return 0
}

// getTenantIDFromContext 从 gin context 中提取租户 ID
func getTenantIDFromContext(c *gin.Context) int64 {
	if v, exists := c.Get("tenantId"); exists {
		return toInt64(v)
	}
	if v, exists := c.Get("tenant_id"); exists {
		return toInt64(v)
	}
	return 0
}

// getRolesFromContext 从 gin context 中提取角色列表
func getRolesFromContext(c *gin.Context) []string {
	if v, exists := c.Get("roles"); exists {
		switch val := v.(type) {
		case []string:
			return val
		case string:
			if val == "" {
				return nil
			}
			return strings.Split(val, ",")
		}
	}
	return nil
}

// getPermissionsFromContext 从 gin context 中提取权限列表
func getPermissionsFromContext(c *gin.Context) []string {
	if v, exists := c.Get("permissions"); exists {
		switch val := v.(type) {
		case []string:
			return val
		case string:
			if val == "" {
				return nil
			}
			return strings.Split(val, ",")
		}
	}
	return nil
}

// toInt64 将 interface{} 转换为 int64
func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
