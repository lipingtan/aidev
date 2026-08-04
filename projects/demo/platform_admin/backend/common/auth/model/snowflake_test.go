package model

import (
	"testing"
)

func TestNextID_Unique(t *testing.T) {
	// 生成 1000 个 ID，验证全部唯一
	ids := make(map[int64]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := NextID()
		if id == 0 {
			t.Error("NextID 不应返回 0")
		}
		if _, exists := ids[id]; exists {
			t.Errorf("发现重复 ID: %d（第 %d 次）", id, i)
		}
		ids[id] = struct{}{}
	}
}

func TestNextID_Positive(t *testing.T) {
	id := NextID()
	if id <= 0 {
		t.Errorf("NextID 应返回正整数，got: %d", id)
	}
}

func TestNodeIDFromPodIP_Range(t *testing.T) {
	nodeID := nodeIDFromPodIP()
	if nodeID < 0 || nodeID >= 1024 {
		t.Errorf("NodeID 应在 [0, 1023] 范围内，got: %d", nodeID)
	}
}

func TestNodeIDFromPodIP_Fallback(t *testing.T) {
	// nodeIDFromPodIP 在无法获取非回环 IPv4 时应返回 1（单机兜底）
	// 这里验证函数至少不 panic 且返回合法值
	nodeID := nodeIDFromPodIP()
	if nodeID < 0 || nodeID >= 1024 {
		t.Errorf("NodeID 超出合法范围，got: %d", nodeID)
	}
}
