package discovery

import (
	"log"

	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// ApiMetadata 开发者声明的 API 元数据
type ApiMetadata struct {
	Name           string // 接口名称
	PermissionCode string // 权限标识码
	URLPattern     string // URL 匹配模式
	HTTPMethod     string // HTTP 方法
	AppCode        string // 所属应用编码
	GroupName      string // 所属分组名（可选，自动创建 GROUP）
}

// RegisterAPIs 代码声明注册 API，状态为 ACTIVE，支持自动创建分组
// appCode: 注册到指定应用
func RegisterAPIs(db *gorm.DB, appCode string, apis []ApiMetadata) {
	if len(apis) == 0 {
		return
	}

	// 查询已有 ENDPOINT，用于幂等去重
	var existing []model.ApiPermission
	if err := db.Where("app_code = ? AND type = ?", appCode, "ENDPOINT").Find(&existing).Error; err != nil {
		log.Printf("[discovery] RegisterAPIs 查询已有 endpoint 失败: %v", err)
		return
	}
	existingMap := make(map[string]struct{}, len(existing))
	for _, ep := range existing {
		existingMap[ep.HTTPMethod+":"+ep.URLPattern] = struct{}{}
	}

	// 查询已有 GROUP，用于避免重复创建分组
	var existingGroups []model.ApiPermission
	if err := db.Where("app_code = ? AND type = ?", appCode, "GROUP").Find(&existingGroups).Error; err != nil {
		log.Printf("[discovery] RegisterAPIs 查询已有 group 失败: %v", err)
		return
	}
	groupMap := make(map[string]int64, len(existingGroups))
	for _, g := range existingGroups {
		groupMap[g.Name] = g.ID
	}

	for _, api := range apis {
		key := api.HTTPMethod + ":" + api.URLPattern
		if _, exists := existingMap[key]; exists {
			continue
		}

		// 处理分组
		var parentID *int64
		if api.GroupName != "" {
			gid, ok := groupMap[api.GroupName]
			if !ok {
				// 自动创建分组
				group := model.ApiPermission{
					Type:    "GROUP",
					Name:    api.GroupName,
					AppCode: api.AppCode,
					Status:  "ACTIVE",
				}
				if err := db.Create(&group).Error; err != nil {
					log.Printf("[discovery] RegisterAPIs 创建分组 %s 失败: %v", api.GroupName, err)
					continue
				}
				gid = group.ID
				groupMap[api.GroupName] = gid
			}
			parentID = &gid
		}

		endpoint := model.ApiPermission{
			ParentID:       parentID,
			Type:           "ENDPOINT",
			Name:           api.Name,
			PermissionCode: api.PermissionCode,
			URLPattern:     api.URLPattern,
			HTTPMethod:     api.HTTPMethod,
			AppCode:        api.AppCode,
			Status:         "ACTIVE",
		}
		if err := db.Create(&endpoint).Error; err != nil {
			log.Printf("[discovery] RegisterAPIs 插入 endpoint %s 失败: %v", api.Name, err)
			continue
		}
		// 加入 map 防止同批次重复
		existingMap[key] = struct{}{}
	}
}
