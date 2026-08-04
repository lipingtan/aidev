package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// ApplyColPolicy 对响应数据应用列权限（HIDE/MASK）。
// data 可以是任意 struct 或 slice，通过 JSON 序列化/反序列化处理。
// colEffects 是 MergeColPolicies 的输出，key 为 JSON 字段名。
// 返回处理后的 interface{}（可直接序列化输出）。
func ApplyColPolicy(data interface{}, colEffects map[string]PolicyColEntry) interface{} {
	if len(colEffects) == 0 {
		return data
	}

	// 序列化为通用 map 结构
	raw, err := json.Marshal(data)
	if err != nil {
		log.Printf("[abac] ApplyColPolicy marshal error: %v", err)
		return data
	}

	// 尝试反序列化为 slice
	var sliceResult []map[string]interface{}
	if err := json.Unmarshal(raw, &sliceResult); err == nil {
		for i := range sliceResult {
			applyToMap(sliceResult[i], colEffects)
		}
		return sliceResult
	}

	// 尝试反序列化为单个 map
	var mapResult map[string]interface{}
	if err := json.Unmarshal(raw, &mapResult); err == nil {
		applyToMap(mapResult, colEffects)
		return mapResult
	}

	// 无法处理，原样返回
	return data
}

// applyToMap 对单个 map 执行列权限处理
func applyToMap(m map[string]interface{}, colEffects map[string]PolicyColEntry) {
	for fieldName, entry := range colEffects {
		if _, exists := m[fieldName]; !exists {
			continue
		}
		switch entry.Effect {
		case ColEffectHide:
			m[fieldName] = nil
		case ColEffectMask:
			m[fieldName] = applyMask(fmt.Sprintf("%v", m[fieldName]), entry.MaskType, entry.MaskPattern)
		// ColEffectShow: 不处理，原样保留
		}
	}
}

// applyMask 对字符串值执行脱敏处理
func applyMask(value string, maskType string, maskPattern string) string {
	switch maskType {
	case "phone":
		return maskPhone(value)
	case "email":
		return maskEmail(value)
	case "id_card":
		return maskIDCard(value)
	case "custom":
		return maskCustom(value, maskPattern)
	default:
		return "***"
	}
}

// maskPhone 手机号脱敏：保留前3后4，中间替换为 ****
func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// maskEmail 邮箱脱敏：用户名部分替换为 ***，保留域名
func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***"
	}
	return "***@" + parts[1]
}

// maskIDCard 身份证脱敏：保留前6后4，中间替换为 ********
func maskIDCard(idCard string) string {
	idCard = strings.TrimSpace(idCard)
	if len(idCard) < 10 {
		return "***"
	}
	return idCard[:6] + "********" + idCard[len(idCard)-4:]
}

// maskCustom 自定义正则脱敏：用 *** 替换匹配部分
func maskCustom(value string, pattern string) string {
	if pattern == "" {
		return "***"
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("[abac] maskCustom invalid pattern %q: %v", pattern, err)
		return value // 编译失败保留原值
	}
	return re.ReplaceAllString(value, "***")
}
