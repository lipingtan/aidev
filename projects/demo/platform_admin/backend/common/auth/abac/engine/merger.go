package engine

// RowMergeResult 行权限合并结果
type RowMergeResult struct {
	Deny       bool             // true 时直接注入 WHERE 1=0
	Conditions []TranslateResult // ALLOW 合并后的 SQL 条件列表（OR 语义）
}

// PolicyRowEntry 用于合并的行权限条目（从 Service 层传入）
type PolicyRowEntry struct {
	Effect        string    // ALLOW | DENY
	Action        string    // read|create|update|delete
	ConditionNode *CondNode // 条件表达式树，nil 表示无条件（match all）
}

// PolicyColEntry 用于合并的列权限条目
type PolicyColEntry struct {
	FieldName   string
	Effect      string // SHOW|HIDE|MASK
	MaskType    string
	MaskPattern string
}

// MergeRowPolicies 合并行权限：DENY 优先，多 ALLOW 取 OR 并集。
// action 为空时不过滤 action，返回所有命中条件。
func MergeRowPolicies(
	rows []PolicyRowEntry,
	action string,
	authInfo map[string]interface{},
	resDef *ResourceDef,
) RowMergeResult {
	var allowConds []TranslateResult
	hasMatchingPolicy := false

	for _, row := range rows {
		// action 过滤
		if action != "" && row.Action != action {
			continue
		}
		hasMatchingPolicy = true

		if row.Effect == EffectDeny {
			// DENY 优先，立即返回（无论 ConditionNode 是否为 nil）
			return RowMergeResult{Deny: true}
		}
		if row.Effect == EffectAllow {
			if row.ConditionNode == nil {
				// 无条件 ALLOW：不注入任何 WHERE（match all）
				// 但需继续遍历检查后续是否有 DENY
				allowConds = append(allowConds, TranslateResult{SQL: ""}) // 标记 match-all
				continue
			}
			result := Translate(row.ConditionNode, authInfo, resDef)
			allowConds = append(allowConds, result)
		}
	}

	_ = hasMatchingPolicy

	// 过滤掉 match-all 标记（SQL=""）：只要有一个 match-all ALLOW，不注入 WHERE
	for _, c := range allowConds {
		if c.SQL == "" {
			return RowMergeResult{Deny: false, Conditions: nil}
		}
	}

	return RowMergeResult{
		Deny:       false,
		Conditions: allowConds,
	}
}

// MergeColPolicies 合并列权限：HIDE > MASK > SHOW，取最严格效果。
// 返回 map[fieldName]PolicyColEntry。
func MergeColPolicies(cols []PolicyColEntry) map[string]PolicyColEntry {
	result := make(map[string]PolicyColEntry)
	for _, c := range cols {
		existing, exists := result[c.FieldName]
		if !exists {
			result[c.FieldName] = c
			continue
		}
		// 合并优先级：HIDE > MASK > SHOW
		if colEffectPriority(c.Effect) > colEffectPriority(existing.Effect) {
			result[c.FieldName] = c
		}
	}
	return result
}

// colEffectPriority 返回列权限效果的优先级数值（越大越严格）
func colEffectPriority(effect string) int {
	switch effect {
	case ColEffectHide:
		return 3
	case ColEffectMask:
		return 2
	case ColEffectShow:
		return 1
	default:
		return 0
	}
}
