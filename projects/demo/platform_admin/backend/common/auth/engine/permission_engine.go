package engine

import "strings"

// PermissionEngine 权限表达式引擎
// 精确码使用 HashSet O(1) 查找，通配符模式 O(n) 遍历匹配
type PermissionEngine struct {
	matchAll         bool                // 是否有 "**" 规则（匹配一切）
	exactSet         map[string]struct{} // 精确码集合
	wildcardPatterns []compiledPattern   // 通配符模式列表
}

// NewPermissionEngine 构造权限引擎，预编译所有权限码模式
func NewPermissionEngine(codes []string) *PermissionEngine {
	e := &PermissionEngine{
		exactSet: make(map[string]struct{}),
	}

	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}

		// "**" 单独作为权限码表示匹配一切
		if code == "**" {
			e.matchAll = true
			continue
		}

		// 判断是否包含通配符
		if strings.Contains(code, "*") {
			p := compilePattern(code)
			e.wildcardPatterns = append(e.wildcardPatterns, p)
		} else {
			e.exactSet[code] = struct{}{}
		}
	}

	return e
}

// HasPermission 检查是否拥有指定权限
func (e *PermissionEngine) HasPermission(required string) bool {
	if e.matchAll {
		return true
	}

	// 精确匹配 O(1)
	if _, ok := e.exactSet[required]; ok {
		return true
	}

	// 通配符匹配 O(n)
	for i := range e.wildcardPatterns {
		if e.wildcardPatterns[i].match(required) {
			return true
		}
	}

	return false
}
