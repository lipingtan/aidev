package engine

import (
	"fmt"
	"log"
	"strings"
)

// TranslateResult 翻译结果
type TranslateResult struct {
	SQL  string
	Args []interface{}
}

// Translate 将条件树节点翻译为 SQL 片段和参数列表。
// 遇到未注册属性名时安全降级返回 "1 = 0"，不 panic，不返回 error。
func Translate(node *CondNode, authInfo map[string]interface{}, resDef *ResourceDef) TranslateResult {
	if node == nil {
		return TranslateResult{SQL: "1 = 1"}
	}

	switch node.Type {
	case "group":
		return translateGroup(node, authInfo, resDef)
	case "expr":
		return translateExpr(node, authInfo, resDef)
	default:
		log.Printf("[abac] unknown condition node type: %s", node.Type)
		return TranslateResult{SQL: "1 = 0"}
	}
}

// translateGroup 翻译组合节点（AND/OR）
func translateGroup(node *CondNode, authInfo map[string]interface{}, resDef *ResourceDef) TranslateResult {
	if len(node.Children) == 0 {
		return TranslateResult{SQL: "1 = 1"}
	}

	op := strings.ToUpper(node.Operator)
	if op != "AND" && op != "OR" {
		op = "AND"
	}

	parts := make([]string, 0, len(node.Children))
	var allArgs []interface{}

	for _, child := range node.Children {
		r := Translate(child, authInfo, resDef)
		parts = append(parts, "("+r.SQL+")")
		allArgs = append(allArgs, r.Args...)
	}

	return TranslateResult{
		SQL:  strings.Join(parts, " "+op+" "),
		Args: allArgs,
	}
}

// translateExpr 翻译叶子节点（属性比较表达式）
func translateExpr(node *CondNode, authInfo map[string]interface{}, resDef *ResourceDef) TranslateResult {
	if node.Left == nil {
		return TranslateResult{SQL: "1 = 0"}
	}

	// 解析左值（必须是资源属性）
	if node.Left.Source != SourceResource {
		log.Printf("[abac] expr left source must be 'resource', got: %s", node.Left.Source)
		return TranslateResult{SQL: "1 = 0"}
	}

	leftAttr, ok := resDef.GetAttr(node.Left.Attr)
	if !ok {
		log.Printf("[abac] unknown resource attr: %s on resource %s", node.Left.Attr, resDef.Type)
		return TranslateResult{SQL: "1 = 0"}
	}

	// 两列互比：右值也是资源属性
	if node.Right != nil && node.Right.Source == SourceResource {
		rightAttr, ok2 := resDef.GetAttr(node.Right.Attr)
		if !ok2 {
			log.Printf("[abac] unknown resource attr (right): %s on resource %s", node.Right.Attr, resDef.Type)
			return TranslateResult{SQL: "1 = 0"}
		}
		sqlOp := opToSQL(node.Op)
		if sqlOp == "" {
			return TranslateResult{SQL: "1 = 0"}
		}
		return TranslateResult{
			SQL:  fmt.Sprintf("%s %s %s", leftAttr.QualifiedCol, sqlOp, rightAttr.QualifiedCol),
			Args: nil,
		}
	}

	// 带 JoinPath：生成 EXISTS 子查询
	if leftAttr.JoinPath != "" {
		return translateJoinPath(leftAttr, node, authInfo, resDef)
	}

	// 普通比较：右值为主体属性或常量
	rightVal := resolveRight(node.Right, authInfo)

	return buildWhere(leftAttr.QualifiedCol, node.Op, rightVal)
}

// translateJoinPath 根据 JoinPath 模板生成 EXISTS 子查询
func translateJoinPath(attr ResourceAttrDef, node *CondNode, authInfo map[string]interface{}, resDef *ResourceDef) TranslateResult {
	rightVal := resolveRight(node.Right, authInfo)
	sqlOp := opToSQL(node.Op)
	if sqlOp == "" {
		return TranslateResult{SQL: "1 = 0"}
	}

	// JoinPath 模板替换占位符
	subQuery := attr.JoinPath
	subQuery = strings.ReplaceAll(subQuery, "{main_table}", resDef.MainTable)
	subQuery = strings.ReplaceAll(subQuery, "{col}", attr.QualifiedCol)
	subQuery = strings.ReplaceAll(subQuery, "{op}", sqlOp)
	subQuery = strings.ReplaceAll(subQuery, "{param}", "?")

	return TranslateResult{
		SQL:  "EXISTS (" + subQuery + ")",
		Args: []interface{}{rightVal},
	}
}

// resolveRight 解析右值为运行时值
func resolveRight(right *CondValue, authInfo map[string]interface{}) interface{} {
	if right == nil {
		return nil
	}
	switch right.Source {
	case SourceConst:
		return right.Value
	case SourceSubject:
		if authInfo == nil {
			return nil
		}
		attrDef, ok := GetSubjectAttr(right.Attr)
		if !ok {
			log.Printf("[abac] unknown subject attr: %s", right.Attr)
			return nil
		}
		return attrDef.Resolver(authInfo)
	default:
		return right.Value
	}
}

// buildWhere 根据列名、操作符、右值生成 SQL 片段
func buildWhere(col string, op string, rightVal interface{}) TranslateResult {
	switch op {
	case OpIsNull:
		return TranslateResult{SQL: col + " IS NULL"}
	case OpIsNotNull:
		return TranslateResult{SQL: col + " IS NOT NULL"}
	case OpIn:
		return TranslateResult{SQL: col + " IN (?)", Args: []interface{}{rightVal}}
	case OpNotIn:
		return TranslateResult{SQL: col + " NOT IN (?)", Args: []interface{}{rightVal}}
	case OpContains:
		return TranslateResult{SQL: col + " LIKE ?", Args: []interface{}{"%" + fmt.Sprintf("%v", rightVal) + "%"}}
	default:
		sqlOp := opToSQL(op)
		if sqlOp == "" {
			log.Printf("[abac] unsupported op: %s", op)
			return TranslateResult{SQL: "1 = 0"}
		}
		return TranslateResult{SQL: col + " " + sqlOp + " ?", Args: []interface{}{rightVal}}
	}
}

// opToSQL 将操作符常量转换为 SQL 运算符
func opToSQL(op string) string {
	switch op {
	case OpEq:
		return "="
	case OpNe:
		return "!="
	case OpGt:
		return ">"
	case OpLt:
		return "<"
	case OpGte:
		return ">="
	case OpLte:
		return "<="
	default:
		return ""
	}
}
