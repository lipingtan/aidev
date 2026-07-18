package service

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// StringInt64Slice 支持 JSON 中的字符串或数字数组，统一解析为 int64 切片
// 前端使用字符串 ID（避免雪花 ID 精度丢失），后端需要 int64
type StringInt64Slice []int64

func (s *StringInt64Slice) UnmarshalJSON(data []byte) error {
	// 尝试解析为 []int64
	var intSlice []int64
	if err := json.Unmarshal(data, &intSlice); err == nil {
		*s = intSlice
		return nil
	}
	// 尝试解析为 []string
	var strSlice []string
	if err := json.Unmarshal(data, &strSlice); err != nil {
		return fmt.Errorf("无法解析 ID 数组: %w", err)
	}
	result := make([]int64, 0, len(strSlice))
	for _, str := range strSlice {
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("无法将 %q 解析为 int64: %w", str, err)
		}
		result = append(result, v)
	}
	*s = result
	return nil
}
