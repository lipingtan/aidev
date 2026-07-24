package service

import (
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ResourceNode 菜单树节点
type ResourceNode struct {
	ID             int64           `json:"id,string"`
	ParentID       *int64          `json:"parent_id,string"`
	Type           string          `json:"type"`
	Name           string          `json:"name"`
	PermissionCode string          `json:"permission_code"`
	Path           string          `json:"path"`
	Component      string          `json:"component"`
	Icon           string          `json:"icon"`
	AppCode        string          `json:"app_code"`
	Platform       string          `json:"platform"`
	ModuleCode     string          `json:"module_code"`
	SortOrder      int             `json:"sort_order"`
	Status         int             `json:"status"`
	Children       []*ResourceNode `json:"children"`
}

// menuCacheEntry 菜单缓存条目
type menuCacheEntry struct {
	tree     []*ResourceNode
	cachedAt time.Time
}

const menuCacheTTL = 10 * time.Minute

// UserMenuService C端用户菜单服务，按 tenant_id 维度缓存
type UserMenuService struct {
	db    *gorm.DB
	cache sync.Map // key: tenantID(int64), value: *menuCacheEntry
}

// NewUserMenuService 创建 UserMenuService 实例
func NewUserMenuService(db *gorm.DB) *UserMenuService {
	return &UserMenuService{db: db}
}

// resourceRow 数据库查询行映射
type resourceRow struct {
	ID             int64
	ParentID       *int64
	Type           string
	Name           string
	PermissionCode string
	Path           string
	Component      string
	Icon           string
	AppCode        string
	Platform       string
	ModuleCode     string
	SortOrder      int
	Status         int
}

// GetUserMenu 获取C端用户菜单
// 逻辑：查 admin_resource 表 WHERE platform='user' AND status=1
// 按 tenant_id 关联 admin_tenant_app 查已启用应用，再按 enabled_modules 过滤
// 返回树形结构，结果按 tenant_id 缓存
func (s *UserMenuService) GetUserMenu(tenantID int64) ([]*ResourceNode, error) {
	// 尝试从缓存获取
	if val, ok := s.cache.Load(tenantID); ok {
		entry := val.(*menuCacheEntry)
		if time.Since(entry.cachedAt) < menuCacheTTL {
			return entry.tree, nil
		}
		// 过期，删除
		s.cache.Delete(tenantID)
	}

	// 缓存 miss，查库
	tree, err := s.queryUserMenu(tenantID)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	s.cache.Store(tenantID, &menuCacheEntry{
		tree:     tree,
		cachedAt: time.Now(),
	})
	return tree, nil
}

// InvalidateUserMenuCache 失效指定租户的菜单缓存
func (s *UserMenuService) InvalidateUserMenuCache(tenantID int64) {
	s.cache.Delete(tenantID)
}

// queryUserMenu 从数据库查询指定租户的 C 端菜单
func (s *UserMenuService) queryUserMenu(tenantID int64) ([]*ResourceNode, error) {
	// 查询该租户订阅的应用列表
	type tenantAppRow struct {
		AppCode        string
		EnabledModules *string
	}
	var tenantApps []tenantAppRow
	err := s.db.Table("admin_tenant_app").
		Select("app_code, enabled_modules").
		Where("tenant_id = ?", tenantID).
		Scan(&tenantApps).Error
	if err != nil {
		return nil, err
	}

	// 无订阅应用则返回空
	if len(tenantApps) == 0 {
		return []*ResourceNode{}, nil
	}

	// 收集 app_code 列表
	appCodes := make([]string, 0, len(tenantApps))
	for _, ta := range tenantApps {
		appCodes = append(appCodes, ta.AppCode)
	}

	// 查询 platform=user AND status=1 的菜单资源
	var rows []resourceRow
	query := s.db.Table("admin_resource").
		Select("id, parent_id, type, name, permission_code, path, component, icon, app_code, platform, module_code, sort_order, status").
		Where("platform = ? AND status = ? AND app_code IN ? AND deleted_at IS NULL", "user", 1, appCodes).
		Order("sort_order ASC, id ASC")

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 构建 enabled_modules 过滤映射: app_code → []module
	appModules := make(map[string][]string)
	for _, ta := range tenantApps {
		if ta.EnabledModules != nil && *ta.EnabledModules != "" {
			modules := parseJSONStringArray(*ta.EnabledModules)
			if len(modules) > 0 {
				appModules[ta.AppCode] = modules
			}
		}
	}

	// 过滤：如果该应用有 enabled_modules 限制，且资源有 module_code，则需在启用列表中
	filtered := make([]resourceRow, 0, len(rows))
	for _, r := range rows {
		if modules, hasLimit := appModules[r.AppCode]; hasLimit && r.ModuleCode != "" {
			if !sliceContains(modules, r.ModuleCode) {
				continue
			}
		}
		filtered = append(filtered, r)
	}

	return buildMenuTree(filtered), nil
}

// buildMenuTree 构建菜单树
func buildMenuTree(rows []resourceRow) []*ResourceNode {
	if len(rows) == 0 {
		return []*ResourceNode{}
	}

	nodeMap := make(map[int64]*ResourceNode, len(rows))
	for _, r := range rows {
		nodeMap[r.ID] = &ResourceNode{
			ID:             r.ID,
			ParentID:       r.ParentID,
			Type:           r.Type,
			Name:           r.Name,
			PermissionCode: r.PermissionCode,
			Path:           r.Path,
			Component:      r.Component,
			Icon:           r.Icon,
			AppCode:        r.AppCode,
			Platform:       r.Platform,
			ModuleCode:     r.ModuleCode,
			SortOrder:      r.SortOrder,
			Status:         r.Status,
		}
	}

	var roots []*ResourceNode
	for _, r := range rows {
		node := nodeMap[r.ID]
		if r.ParentID == nil || *r.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*r.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 父节点不在结果集中，作为根节点
			roots = append(roots, node)
		}
	}

	if roots == nil {
		return []*ResourceNode{}
	}
	return roots
}

// parseJSONStringArray 简单解析 JSON 字符串数组 ["a","b","c"]
func parseJSONStringArray(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" || s == "[]" {
		return nil
	}
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// sliceContains 判断字符串切片中是否包含目标字符串
func sliceContains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
