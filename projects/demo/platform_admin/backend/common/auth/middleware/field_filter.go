package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FieldObjectRegistry 路由-对象映射注册表
// 存储 method:path → objectCode 映射，用于 FieldFilterMiddleware 匹配请求对应的业务对象
type FieldObjectRegistry struct {
	mu     sync.RWMutex
	routes map[string]string // "GET:/api/v1/admin/users" → "user"
}

// NewFieldObjectRegistry 创建空的路由-对象映射注册表
func NewFieldObjectRegistry() *FieldObjectRegistry {
	return &FieldObjectRegistry{
		routes: make(map[string]string),
	}
}

// RegisterRoute 注册路由与 objectCode 的映射
func (r *FieldObjectRegistry) RegisterRoute(method, path, objectCode string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.ToUpper(method) + ":" + path
	r.routes[key] = objectCode
}

// MatchRoute 根据请求的 method 和 path 匹配 objectCode
// path 应使用 c.FullPath() 获取（含路径参数模板，如 /users/:id）
func (r *FieldObjectRegistry) MatchRoute(method, path string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := strings.ToUpper(method) + ":" + path
	return r.routes[key]
}

// LoadRoutes 从数据库 admin_field_object 表加载对象列表（objectCode），
// 但路由映射需要硬编码或通过其他配置提供，此方法仅作为扩展预留
func (r *FieldObjectRegistry) LoadRoutes(db *gorm.DB) {
	// 当前版本从 admin_field_object 加载已注册的对象，
	// 路由映射由代码硬编码注册（RegisterRoute）
	var objects []model.FieldObject
	if err := db.Find(&objects).Error; err != nil {
		log.Printf("[field-filter] 加载字段对象列表失败: %v", err)
		return
	}
	log.Printf("[field-filter] 已加载 %d 个字段对象", len(objects))
}

// accessLevel 权限级别，用于冲突解决时比较优先级
type accessLevel int

const (
	accessHidden   accessLevel = 0
	accessVisible  accessLevel = 1
	accessEditable accessLevel = 2
)

// parseAccessLevel 解析权限级别字符串为枚举值
func parseAccessLevel(s string) accessLevel {
	switch strings.ToUpper(s) {
	case "EDITABLE":
		return accessEditable
	case "VISIBLE":
		return accessVisible
	default:
		return accessHidden
	}
}

// fieldFilterWriter 自定义 ResponseWriter，用于捕获 Handler 写入的响应体
type fieldFilterWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 捕获写入的数据到缓冲区
func (w *fieldFilterWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

// WriteString 捕获写入的字符串到缓冲区
func (w *fieldFilterWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// FieldFilterMiddleware 字段权限过滤中间件
// 自动拦截 JSON 响应，根据路由注册的 objectCode 过滤 HIDDEN 字段
func FieldFilterMiddleware(registry *FieldObjectRegistry, permRepo repository.FieldPermissionRepository, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 替换 ResponseWriter 以捕获响应体
		blw := &fieldFilterWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = blw

		// 执行后续 Handler
		c.Next()

		// 检查是否被手动标记为跳过
		if c.GetBool("skip_field_filter") {
			// 跳过过滤，直接写入原始响应
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// 根据路由路径匹配 objectCode
		objectCode := registry.MatchRoute(c.Request.Method, c.FullPath())
		if objectCode == "" {
			// 未注册的路由，直接写入原始响应
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// 获取用户认证上下文
		authCtx := GetAuthContext(c)
		if authCtx == nil || len(authCtx.Roles) == 0 {
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// 查询字段权限配置
		perms, err := permRepo.GetByRolesAndObject(db, authCtx.Roles, objectCode)
		if err != nil {
			log.Printf("[field-filter] 查询字段权限失败: %v", err)
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// 未配置字段权限时不做任何过滤（RG-3）
		if len(perms) == 0 {
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// 计算 HIDDEN 字段集合（多角色冲突取最高权限）
		hiddenFields := resolveHiddenFields(perms)
		if len(hiddenFields) == 0 {
			// 无需隐藏任何字段
			blw.ResponseWriter.Write(blw.body.Bytes())
			return
		}

		// 过滤响应体中的 HIDDEN 字段
		filtered := filterJSON(blw.body.Bytes(), hiddenFields)
		blw.ResponseWriter.Write(filtered)
	}
}

// resolveHiddenFields 解析多角色字段权限，返回需要隐藏的字段名集合
// 冲突解决规则：多角色取最高权限（EDITABLE > VISIBLE > HIDDEN）
func resolveHiddenFields(perms []model.FieldPermission) map[string]struct{} {
	// 合并多角色权限，每个字段取最高权限级别
	fieldMaxLevel := make(map[string]accessLevel)
	for _, p := range perms {
		level := parseAccessLevel(p.Access)
		if existing, ok := fieldMaxLevel[p.FieldName]; !ok || level > existing {
			fieldMaxLevel[p.FieldName] = level
		}
	}

	// 收集最终权限为 HIDDEN 的字段
	hidden := make(map[string]struct{})
	for field, level := range fieldMaxLevel {
		if level == accessHidden {
			hidden[field] = struct{}{}
		}
	}
	return hidden
}

// filterJSON 从 JSON 响应体中移除 HIDDEN 字段（仅处理顶层 data 字段）
// 支持标准响应格式 {"code":0,"data":{...},"message":"ok"} 和 {"code":0,"data":[{...}],"message":"ok"}
func filterJSON(body []byte, hiddenFields map[string]struct{}) []byte {
	if len(body) == 0 {
		return body
	}

	// 解析 JSON
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		// 非 JSON 响应，原样返回
		return body
	}

	// 获取 data 字段
	data, exists := response["data"]
	if !exists || data == nil {
		return body
	}

	// 根据 data 类型过滤字段
	switch v := data.(type) {
	case map[string]interface{}:
		// data 是单个对象 — 检查是否有 list 字段（分页响应）
		if list, ok := v["list"]; ok {
			if items, ok := list.([]interface{}); ok {
				// 分页响应：过滤 list 中每个对象
				for i, item := range items {
					if obj, ok := item.(map[string]interface{}); ok {
						items[i] = removeFields(obj, hiddenFields)
					}
				}
				v["list"] = items
			}
		} else {
			// 单个对象响应
			response["data"] = removeFields(v, hiddenFields)
		}
	case []interface{}:
		// data 是数组
		for i, item := range v {
			if obj, ok := item.(map[string]interface{}); ok {
				v[i] = removeFields(obj, hiddenFields)
			}
		}
		response["data"] = v
	}

	// 重新序列化
	result, err := json.Marshal(response)
	if err != nil {
		return body
	}
	return result
}

// removeFields 从 map 中移除指定字段（仅顶层）
func removeFields(obj map[string]interface{}, hiddenFields map[string]struct{}) map[string]interface{} {
	for field := range hiddenFields {
		delete(obj, field)
	}
	return obj
}
