package spi

import "gorm.io/gorm"

// DefaultOrganizationProvider 基于 admin_org_unit + admin_user_org 的默认组织架构提供者
type DefaultOrganizationProvider struct {
	DB *gorm.DB
}

// GetOrgIds 获取用户在该租户下的主归属组织节点 ID（is_primary=1）
func (p *DefaultOrganizationProvider) GetOrgIds(userID, tenantID int64) ([]int64, error) {
	var orgIDs []int64
	err := p.DB.Table("admin_user_org").
		Where("user_id = ? AND tenant_id = ? AND is_primary = 1", userID, tenantID).
		Pluck("org_unit_id", &orgIDs).Error
	if err != nil {
		return []int64{}, nil
	}
	return orgIDs, nil
}

// GetOrgPath 获取从指定节点到根的路径
func (p *DefaultOrganizationProvider) GetOrgPath(orgID, tenantID int64) ([]int64, error) {
	var path []int64
	current := orgID
	for current != 0 {
		path = append(path, current)
		var parentID *int64
		result := p.DB.Table("admin_org_unit").
			Where("id = ? AND tenant_id = ?", current, tenantID).
			Pluck("parent_id", &parentID)
		if result.Error != nil || parentID == nil {
			break
		}
		current = *parentID
	}
	if path == nil {
		return []int64{}, nil
	}
	return path, nil
}

// GetSubOrgIds 获取指定节点及所有子节点 ID（单次查询 + 内存 BFS）
func (p *DefaultOrganizationProvider) GetSubOrgIds(orgID, tenantID int64) ([]int64, error) {
	// 单次查询该租户所有启用节点
	type node struct {
		ID       int64 `gorm:"column:id"`
		ParentID int64 `gorm:"column:parent_id"`
	}
	var allNodes []node
	err := p.DB.Table("admin_org_unit").
		Select("id, COALESCE(parent_id, 0) as parent_id").
		Where("tenant_id = ? AND status = 1", tenantID).
		Find(&allNodes).Error
	if err != nil || len(allNodes) == 0 {
		return []int64{}, nil
	}

	// 内存构建父子映射
	childMap := make(map[int64][]int64)
	for _, n := range allNodes {
		childMap[n.ParentID] = append(childMap[n.ParentID], n.ID)
	}

	// BFS 从 orgID 开始收集子树
	var result []int64
	queue := []int64{orgID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		queue = append(queue, childMap[current]...)
	}
	return result, nil
}
