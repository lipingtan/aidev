package plugin

import (
	"encoding/json"
	"fmt"

	"go-admin/common/auth/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PluginResourceSyncer 负责将插件声明的资源同步到 RBAC 体系
type PluginResourceSyncer struct {
	db *gorm.DB
}

// NewPluginResourceSyncer 创建资源同步器
func NewPluginResourceSyncer(db *gorm.DB) *PluginResourceSyncer {
	return &PluginResourceSyncer{db: db}
}

// SyncOnStart 插件启动时全量同步资源到 admin_resource / admin_api_permission
// 在单事务中执行：硬删除旧资源 → 批量写入新资源 → 清理孤儿绑定 → 更新应用元数据
func (s *PluginResourceSyncer) SyncOnStart(appCode string, manifest *ManifestV2) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 硬删除该 app_code 的旧资源（Unscoped 绕过软删除）
		if err := tx.Unscoped().Where("app_code = ?", appCode).Delete(&model.Resource{}).Error; err != nil {
			return fmt.Errorf("清理旧 resource 失败: %w", err)
		}
		if err := tx.Unscoped().Where("app_code = ?", appCode).Delete(&model.ApiPermission{}).Error; err != nil {
			return fmt.Errorf("清理旧 api_permission 失败: %w", err)
		}

		// 2. 批量写入新菜单资源
		if len(manifest.Menus) > 0 {
			resources := s.buildResources(manifest.Menus, appCode, nil)
			if len(resources) > 0 {
				if err := tx.Create(&resources).Error; err != nil {
					return fmt.Errorf("写入 resource 失败: %w", err)
				}
			}
		}

		// 3. 批量写入新 API 权限
		if len(manifest.ApiPermissions) > 0 {
			apiPerms := s.buildApiPermissions(manifest.ApiPermissions, appCode, nil)
			if len(apiPerms) > 0 {
				if err := tx.Create(&apiPerms).Error; err != nil {
					return fmt.Errorf("写入 api_permission 失败: %w", err)
				}
			}
		}

		// 4. 清理孤儿角色绑定（引用了已不存在的 resource_id / api_permission_id）
		if err := s.cleanOrphanBindings(tx, appCode); err != nil {
			return fmt.Errorf("清理孤儿绑定失败: %w", err)
		}

		// 5. 同步更新 admin_application 元数据
		if err := s.syncApplicationMeta(tx, manifest); err != nil {
			return fmt.Errorf("同步应用元数据失败: %w", err)
		}

		return nil
	})
}

// SyncOnUninstall 插件卸载时清理所有资源和关联
func (s *PluginResourceSyncer) SyncOnUninstall(appCode string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 查出要删除的资源 ID
		var resIDs []int64
		tx.Model(&model.Resource{}).Where("app_code = ?", appCode).Pluck("id", &resIDs)
		var apiIDs []int64
		tx.Model(&model.ApiPermission{}).Where("app_code = ?", appCode).Pluck("id", &apiIDs)

		// 级联清理角色绑定
		if len(resIDs) > 0 {
			if err := tx.Where("resource_id IN ?", resIDs).Delete(&model.RoleResource{}).Error; err != nil {
				return fmt.Errorf("清理 role_resource 失败: %w", err)
			}
		}
		if len(apiIDs) > 0 {
			if err := tx.Where("api_permission_id IN ?", apiIDs).Delete(&model.RoleApi{}).Error; err != nil {
				return fmt.Errorf("清理 role_api 失败: %w", err)
			}
		}

		// 清理 role_app 绑定
		if err := tx.Where("app_code = ?", appCode).Delete(&model.RoleApp{}).Error; err != nil {
			return fmt.Errorf("清理 role_app 失败: %w", err)
		}

		// 删除资源和 API 权限
		if err := tx.Unscoped().Where("app_code = ?", appCode).Delete(&model.Resource{}).Error; err != nil {
			return fmt.Errorf("删除 resource 失败: %w", err)
		}
		if err := tx.Unscoped().Where("app_code = ?", appCode).Delete(&model.ApiPermission{}).Error; err != nil {
			return fmt.Errorf("删除 api_permission 失败: %w", err)
		}

		// 软删除 admin_application
		if err := tx.Where("app_code = ?", appCode).Delete(&model.Application{}).Error; err != nil {
			return fmt.Errorf("软删除 application 失败: %w", err)
		}

		// 删除租户订阅关系
		if err := tx.Where("app_code = ?", appCode).Delete(&model.TenantApp{}).Error; err != nil {
			return fmt.Errorf("删除 tenant_app 失败: %w", err)
		}

		return nil
	})
}

// buildResources 递归将 ManifestMenu 转换为 model.Resource 列表
func (s *PluginResourceSyncer) buildResources(menus []ManifestMenu, appCode string, parentID *int64) []model.Resource {
	var resources []model.Resource
	for _, menu := range menus {
		id := model.NextID()
		res := model.Resource{
			ID:             id,
			ParentID:       parentID,
			Type:           menu.Type,
			Name:           menu.Name,
			PermissionCode: menu.PermissionCode,
			Path:           menu.Path,
			Component:      menu.Component,
			Icon:           menu.Icon,
			AppCode:        appCode,
			Platform:       menu.Platform,
			ModuleCode:     menu.ModuleCode,
			SortOrder:      menu.Sort,
			Status:         1,
			Version:        1,
		}
		resources = append(resources, res)

		// 递归处理子菜单
		if len(menu.Children) > 0 {
			children := s.buildResources(menu.Children, appCode, &id)
			resources = append(resources, children...)
		}
	}
	return resources
}

// buildApiPermissions 递归将 ManifestApiPermission 转换为 model.ApiPermission 列表
func (s *PluginResourceSyncer) buildApiPermissions(perms []ManifestApiPermission, appCode string, parentID *int64) []model.ApiPermission {
	var apiPerms []model.ApiPermission
	for _, p := range perms {
		id := model.NextID()
		authRequired := 1
		visible := 1
		ap := model.ApiPermission{
			ID:             id,
			ParentID:       parentID,
			Type:           p.Type,
			Name:           p.Name,
			DisplayName:    p.DisplayName,
			PermissionCode: p.PermissionCode,
			URLPattern:     p.URLPattern,
			HTTPMethod:     p.HTTPMethod,
			AppCode:        appCode,
			ModuleCode:     p.ModuleCode,
			Status:         "ACTIVE",
			Visible:        &visible,
			AuthRequired:   &authRequired,
			SortOrder:      0,
		}
		apiPerms = append(apiPerms, ap)

		// 递归处理子节点
		if len(p.Children) > 0 {
			children := s.buildApiPermissions(p.Children, appCode, &id)
			apiPerms = append(apiPerms, children...)
		}
	}
	return apiPerms
}

// cleanOrphanBindings 清理引用了已不存在 resource/api 的角色绑定
func (s *PluginResourceSyncer) cleanOrphanBindings(tx *gorm.DB, appCode string) error {
	// 获取当前该 app_code 下有效的 resource IDs
	var validResIDs []int64
	tx.Model(&model.Resource{}).Where("app_code = ?", appCode).Pluck("id", &validResIDs)

	// 获取当前该 app_code 下有效的 api_permission IDs
	var validApiIDs []int64
	tx.Model(&model.ApiPermission{}).Where("app_code = ?", appCode).Pluck("id", &validApiIDs)

	// 清理 role_resource 中引用了该 app_code 旧资源（已被删除重建，ID 变了）的绑定
	// 策略：查出所有绑定了该 app 资源但 resource_id 不在新有效列表中的记录
	if len(validResIDs) > 0 {
		// 先查出所有曾绑定该应用资源的 resource_id（通过 app_code 反查）
		// 由于硬删除后旧 ID 已不在 admin_resource 表中，
		// admin_role_resource 中仍然引用旧 ID 的记录就是孤儿
		tx.Exec(`DELETE FROM admin_role_resource WHERE resource_id NOT IN (SELECT id FROM admin_resource) AND resource_id != 0`)
	} else {
		// 该 app 没有新资源，但可能有旧绑定残留（边界情况）
		// 不做全局清理，避免误删其他应用的绑定
	}

	if len(validApiIDs) > 0 {
		tx.Exec(`DELETE FROM admin_role_api WHERE api_permission_id NOT IN (SELECT id FROM admin_api_permission) AND api_permission_id != 0`)
	}

	return nil
}

// syncApplicationMeta 同步更新 admin_application 的元数据
func (s *PluginResourceSyncer) syncApplicationMeta(tx *gorm.DB, manifest *ManifestV2) error {
	platformsJSON, _ := json.Marshal(manifest.Platforms)
	modulesJSON, _ := json.Marshal(manifest.Modules)

	updates := map[string]interface{}{
		"name":         manifest.DisplayName,
		"description":  manifest.Description,
		"route_prefix": manifest.RoutePrefix,
		"platforms":    datatypes.JSON(platformsJSON),
		"modules":      datatypes.JSON(modulesJSON),
	}
	return tx.Model(&model.Application{}).
		Where("app_code = ?", manifest.Name).
		Updates(updates).Error
}
