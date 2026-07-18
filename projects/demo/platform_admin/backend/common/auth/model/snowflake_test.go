package model

import (
	"testing"
)

// TestNextID_Uniqueness 验证生成 1000 个 ID 全部唯一
func TestNextID_Uniqueness(t *testing.T) {
	const count = 1000
	ids := make(map[int64]bool, count)

	for i := 0; i < count; i++ {
		id := NextID()
		if ids[id] {
			t.Fatalf("第 %d 次生成的 ID %d 重复", i, id)
		}
		ids[id] = true
	}
}

// TestNextID_Positive 验证生成的 ID 大于 0
func TestNextID_Positive(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := NextID()
		if id <= 0 {
			t.Fatalf("生成的 ID 应大于 0，实际为 %d", id)
		}
	}
}

// TestNextID_Increasing 验证递增趋势（后生成的 >= 先生成的）
func TestNextID_Increasing(t *testing.T) {
	prev := NextID()
	for i := 0; i < 1000; i++ {
		curr := NextID()
		if curr < prev {
			t.Fatalf("ID 未保持递增趋势: prev=%d, curr=%d", prev, curr)
		}
		prev = curr
	}
}
