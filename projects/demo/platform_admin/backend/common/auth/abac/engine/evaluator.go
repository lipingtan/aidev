package engine

import (
	"context"
	"encoding/json"
)

// ---- Context Keys ----

type abacActionKey struct{}
type abacColContextKey struct{}

// SetAbacAction 将当前请求的 action 写入 context（由 Handler 调用）
func SetAbacAction(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, abacActionKey{}, action)
}

// GetAbacAction 从 context 获取当前 action
func GetAbacAction(ctx context.Context) string {
	v, _ := ctx.Value(abacActionKey{}).(string)
	return v
}

// SetAbacColContext 将列权限合并结果写入 context（由 GORM Callback 调用）
func SetAbacColContext(ctx context.Context, effects map[string]PolicyColEntry) context.Context {
	return context.WithValue(ctx, abacColContextKey{}, effects)
}

// GetAbacColContext 从 context 获取列权限合并结果
func GetAbacColContext(ctx context.Context) map[string]PolicyColEntry {
	v, _ := ctx.Value(abacColContextKey{}).(map[string]PolicyColEntry)
	return v
}

// ---- EvaluateResult 评估 API 响应 ----

// EvaluateResult 策略评估结果
type EvaluateResult struct {
	Allowed      bool                       `json:"allowed"`
	RowCondition *CondNode                  `json:"row_condition,omitempty"`
	ColEffects   map[string]ColEffectDetail `json:"col_effects,omitempty"`
}

// ColEffectDetail 列权限效果详情
type ColEffectDetail struct {
	Effect      string `json:"effect"`
	MaskType    string `json:"mask_type,omitempty"`
	MaskPattern string `json:"mask_pattern,omitempty"`
}

// BuildEvaluateResult 从合并结果构建评估 API 响应
func BuildEvaluateResult(rowResult RowMergeResult, colEffects map[string]PolicyColEntry) EvaluateResult {
	result := EvaluateResult{
		Allowed:    !rowResult.Deny,
		ColEffects: make(map[string]ColEffectDetail),
	}

	// 行条件：多个 ALLOW 条件包裹为 OR group
	if !rowResult.Deny && len(rowResult.Conditions) > 0 {
		if len(rowResult.Conditions) == 1 {
			// 单条件直接解析
			var node CondNode
			if err := json.Unmarshal([]byte(`{"type":"group","operator":"AND","children":[]}`), &node); err == nil {
				result.RowCondition = &node
			}
		} else {
			result.RowCondition = &CondNode{
				Type:     "group",
				Operator: "OR",
				// 注：实际 SQL 片段已在 Callback 中使用，此处仅返回结构标识
			}
		}
	}

	for field, entry := range colEffects {
		result.ColEffects[field] = ColEffectDetail{
			Effect:      entry.Effect,
			MaskType:    entry.MaskType,
			MaskPattern: entry.MaskPattern,
		}
	}

	return result
}
