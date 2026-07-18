// Package engine 权限表达式引擎
package engine

import "strings"

// compiledPattern 预编译的通配符模式
type compiledPattern struct {
	segments    []string // 按冒号分割后的段
	hasGlobStar bool     // 是否含 **
}

// compilePattern 编译权限码模式
func compilePattern(code string) compiledPattern {
	segments := strings.Split(code, ":")
	hasGlobStar := false
	for _, seg := range segments {
		if seg == "**" {
			hasGlobStar = true
			break
		}
	}
	return compiledPattern{
		segments:    segments,
		hasGlobStar: hasGlobStar,
	}
}

// match 判断模式是否匹配目标权限码
func (p *compiledPattern) match(target string) bool {
	targetSegments := strings.Split(target, ":")
	return matchSegments(p.segments, targetSegments)
}

// matchSegments 递归匹配模式段与目标段
// * 匹配恰好一层（不含冒号分隔符）
// ** 匹配一层或多层
func matchSegments(pattern, target []string) bool {
	pi, ti := 0, 0
	for pi < len(pattern) && ti < len(target) {
		switch {
		case pattern[pi] == "**":
			// ** 在末尾，匹配剩余所有层
			if pi == len(pattern)-1 {
				return true
			}
			// 尝试 ** 匹配 1 到 N 层（** 至少匹配一层）
			pi++
			for skip := ti; skip <= len(target); skip++ {
				if matchSegments(pattern[pi:], target[skip:]) {
					return true
				}
			}
			return false
		case pattern[pi] == "*":
			// * 匹配恰好一层
			pi++
			ti++
		case pattern[pi] == target[ti]:
			// 精确匹配
			pi++
			ti++
		default:
			return false
		}
	}
	return pi == len(pattern) && ti == len(target)
}
