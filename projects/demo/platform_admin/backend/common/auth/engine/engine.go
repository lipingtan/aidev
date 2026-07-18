// Package engine 权限表达式引擎
//
// 提供基于通配符模式的权限码匹配能力：
//   - 精确码 O(1) HashSet 查找
//   - * 匹配恰好一层（不含冒号分隔符）
//   - ** 匹配一层或多层
package engine
