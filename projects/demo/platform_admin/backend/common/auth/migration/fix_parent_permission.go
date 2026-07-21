package migration

import (
	"fmt"
	"log"

	"go-admin/common/auth/model"

	"gorm.io/gorm"
)

// TrimmedPermission 被裁剪的权限记录
type TrimmedPermission struct {
	RoleID      int64   `json:"role_id"`
	RoleName    string  `json:"role_name"`
	RemovedRes  []int64 `json:"removed_resource_ids,omitempty"`
	RemovedApis []int64 `json:"removed_api_ids,omitempty"`
}

// FixReport 修复报告
type FixReport struct {
	TotalRolesScanned int                 `json:"total_roles_scanned"`
	AffectedRoles     []TrimmedPermission `json:"affected_roles"`
}

// FixParentPermissionData 一次性数据修复脚本：确保所有子角色权限 ⊆ 父角色权限
// 幂等可重复执行：裁剪后再执行不会有任何变更
func FixParentPermissionData(db *gorm.DB) error {
	log.Println("[FixParentPermission] 开始执行数据修复...")

	// 1. 查询所有有 parent_id 的角色，按层级深度排序（顶层先处理）
	// 通过递归查询获取深度信息比较复杂，这里使用简单方式：
	// 先查所有有 parent_id 的角色，构建层级关系，按深度从浅到深排序处理
	var roles []model.Role
	if err := db.Where("parent_id IS NOT NULL AND role_type != ?", "PERMISSION_SET").
		Order("parent_id ASC").
		Find(&roles).Error; err != nil {
		return fmt.Errorf("查询角色失败: %w", err)
	}

	if len(roles) == 0 {
		log.Println("[FixParentPermission] 未找到有父角色的非 PERMISSION_SET 角色，无需修复")
		return nil
	}

	// 按深度排序：先处理顶层子角色，再处理更深层级
	sortedRoles := sortByDepth(db, roles)

	report := FixReport{
		TotalRolesScanned: len(sortedRoles),
		AffectedRoles:     make([]TrimmedPermission, 0),
	}

	// 2. 在事务中执行裁剪
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, role := range sortedRoles {
			if role.ParentID == nil {
				continue
			}

			trimmed := TrimmedPermission{
				RoleID:   role.ID,
				RoleName: role.RoleName,
			}

			// 3. 检查并裁剪 resourceIDs
			removedRes, err := trimResources(tx, role.ID, *role.ParentID)
			if err != nil {
				return fmt.Errorf("裁剪角色 %d 资源权限失败: %w", role.ID, err)
			}
			trimmed.RemovedRes = removedRes

			// 4. 检查并裁剪 apiIDs
			removedApis, err := trimApis(tx, role.ID, *role.ParentID)
			if err != nil {
				return fmt.Errorf("裁剪角色 %d 接口权限失败: %w", role.ID, err)
			}
			trimmed.RemovedApis = removedApis

			// 只记录有实际裁剪的角色
			if len(trimmed.RemovedRes) > 0 || len(trimmed.RemovedApis) > 0 {
				report.AffectedRoles = append(report.AffectedRoles, trimmed)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("事务执行失败: %w", err)
	}

	// 5. 输出修复报告
	printReport(report)
	return nil
}

// trimResources 裁剪子角色超出父角色范围的资源权限，返回被移除的资源 ID
func trimResources(tx *gorm.DB, childRoleID, parentRoleID int64) ([]int64, error) {
	// 查询父角色的资源 ID
	parentResIDs := getRoleResourceIDs(tx, parentRoleID)
	parentSet := int64ToSet(parentResIDs)

	// 查询子角色的资源 ID
	childResIDs := getRoleResourceIDs(tx, childRoleID)

	// 计算超出部分
	var removed []int64
	var kept []int64
	for _, id := range childResIDs {
		if parentSet[id] {
			kept = append(kept, id)
		} else {
			removed = append(removed, id)
		}
	}

	// 无超出，无需裁剪
	if len(removed) == 0 {
		return nil, nil
	}

	// 删除旧绑定
	if err := tx.Where("role_id = ?", childRoleID).Delete(&model.RoleResource{}).Error; err != nil {
		return nil, err
	}

	// 插入交集部分
	for _, resID := range kept {
		rr := &model.RoleResource{
			RoleID:     childRoleID,
			ResourceID: resID,
		}
		if err := tx.Create(rr).Error; err != nil {
			return nil, err
		}
	}

	return removed, nil
}

// trimApis 裁剪子角色超出父角色范围的接口权限，返回被移除的 API ID
func trimApis(tx *gorm.DB, childRoleID, parentRoleID int64) ([]int64, error) {
	// 查询父角色的 API 权限 ID
	parentApiIDs := getRoleApiIDs(tx, parentRoleID)
	parentSet := int64ToSet(parentApiIDs)

	// 查询子角色的 API 权限 ID
	childApiIDs := getRoleApiIDs(tx, childRoleID)

	// 计算超出部分
	var removed []int64
	var kept []int64
	for _, id := range childApiIDs {
		if parentSet[id] {
			kept = append(kept, id)
		} else {
			removed = append(removed, id)
		}
	}

	// 无超出，无需裁剪
	if len(removed) == 0 {
		return nil, nil
	}

	// 删除旧绑定
	if err := tx.Where("role_id = ?", childRoleID).Delete(&model.RoleApi{}).Error; err != nil {
		return nil, err
	}

	// 插入交集部分
	for _, apiID := range kept {
		ra := &model.RoleApi{
			RoleID:          childRoleID,
			ApiPermissionID: apiID,
		}
		if err := tx.Create(ra).Error; err != nil {
			return nil, err
		}
	}

	return removed, nil
}

// sortByDepth 按层级深度从浅到深排序角色（确保父角色先被处理）
func sortByDepth(db *gorm.DB, roles []model.Role) []model.Role {
	// 构建 ID -> Role 映射
	roleMap := make(map[int64]*model.Role, len(roles))
	for i := range roles {
		roleMap[roles[i].ID] = &roles[i]
	}

	// 计算每个角色的深度（从根到该角色的路径长度）
	depthCache := make(map[int64]int)
	var getDepth func(roleID int64) int
	getDepth = func(roleID int64) int {
		if d, ok := depthCache[roleID]; ok {
			return d
		}
		r, exists := roleMap[roleID]
		if !exists {
			// 不在待处理列表中（可能是顶级角色），查库
			var role model.Role
			if err := db.Where("id = ?", roleID).First(&role).Error; err != nil {
				depthCache[roleID] = 0
				return 0
			}
			if role.ParentID == nil || *role.ParentID == 0 {
				depthCache[roleID] = 0
				return 0
			}
			depth := getDepth(*role.ParentID) + 1
			depthCache[roleID] = depth
			return depth
		}
		if r.ParentID == nil || *r.ParentID == 0 {
			depthCache[roleID] = 0
			return 0
		}
		depth := getDepth(*r.ParentID) + 1
		depthCache[roleID] = depth
		return depth
	}

	// 计算所有角色的深度
	type roleWithDepth struct {
		role  model.Role
		depth int
	}
	rolesWithDepth := make([]roleWithDepth, 0, len(roles))
	for _, r := range roles {
		d := getDepth(r.ID)
		rolesWithDepth = append(rolesWithDepth, roleWithDepth{role: r, depth: d})
	}

	// 按深度排序（浅的先处理）
	for i := 0; i < len(rolesWithDepth)-1; i++ {
		for j := i + 1; j < len(rolesWithDepth); j++ {
			if rolesWithDepth[j].depth < rolesWithDepth[i].depth {
				rolesWithDepth[i], rolesWithDepth[j] = rolesWithDepth[j], rolesWithDepth[i]
			}
		}
	}

	sorted := make([]model.Role, 0, len(rolesWithDepth))
	for _, rwd := range rolesWithDepth {
		sorted = append(sorted, rwd.role)
	}
	return sorted
}

// getRoleResourceIDs 查询角色绑定的资源 ID 列表
func getRoleResourceIDs(db *gorm.DB, roleID int64) []int64 {
	var ids []int64
	db.Model(&model.RoleResource{}).Where("role_id = ?", roleID).Pluck("resource_id", &ids)
	return ids
}

// getRoleApiIDs 查询角色绑定的接口权限 ID 列表
func getRoleApiIDs(db *gorm.DB, roleID int64) []int64 {
	var ids []int64
	db.Model(&model.RoleApi{}).Where("role_id = ?", roleID).Pluck("api_permission_id", &ids)
	return ids
}

// int64ToSet 将 int64 切片转为 map set
func int64ToSet(ids []int64) map[int64]bool {
	m := make(map[int64]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// printReport 打印修复报告
func printReport(report FixReport) {
	log.Printf("[FixParentPermission] 扫描角色总数: %d", report.TotalRolesScanned)

	if len(report.AffectedRoles) == 0 {
		log.Println("[FixParentPermission] 所有子角色权限均在父角色范围内，无需修复")
		return
	}

	log.Printf("[FixParentPermission] 受影响角色数: %d", len(report.AffectedRoles))
	log.Println("[FixParentPermission] ========== 修复报告 ==========")

	for _, tr := range report.AffectedRoles {
		log.Printf("[FixParentPermission] 角色: %s (ID: %d)", tr.RoleName, tr.RoleID)
		if len(tr.RemovedRes) > 0 {
			log.Printf("[FixParentPermission]   - 裁剪资源权限 %d 项, ID: %v", len(tr.RemovedRes), tr.RemovedRes)
		}
		if len(tr.RemovedApis) > 0 {
			log.Printf("[FixParentPermission]   - 裁剪接口权限 %d 项, ID: %v", len(tr.RemovedApis), tr.RemovedApis)
		}
	}

	log.Println("[FixParentPermission] ==============================")
	log.Println("[FixParentPermission] 修复完成")
}
