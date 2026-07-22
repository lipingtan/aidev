package plugin

import (
	"context"
	"errors"
	"testing"
	"time"

	"platform-admin/plugin-sdk/proto"
)

// TestCallPlugin_VersionIncompatible 带 action_version 且 major 不匹配 → ErrActionVersionIncompatible
func TestCallPlugin_VersionIncompatible(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	mock := newMockPlugin("target", "1.0.0", "target")
	_ = mgr.Register("target", mock)
	_ = mgr.Start("target")

	// 手动注册 Action（版本 1.0）
	mgr.ActionRegistry().Register("target", []ManifestAction{
		{Name: "getData", Version: "1.0"},
	})

	bus := mgr.EventBus()

	// 请求版本 2.0，注册版本 1.0 → major 不匹配
	_, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin:  "target",
		Method:        "getData",
		Payload:       []byte(`{}`),
		ActionVersion: "2.0",
	})
	if !errors.Is(err, ErrActionVersionIncompatible) {
		t.Fatalf("期望 ErrActionVersionIncompatible，实际: %v", err)
	}
}

// TestCallPlugin_ActionNotFound 带 action_version 且 action 不存在 → ErrActionNotFound
func TestCallPlugin_ActionNotFound(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	mock := newMockPlugin("target", "1.0.0", "target")
	_ = mgr.Register("target", mock)
	_ = mgr.Start("target")

	// 注册一个不同名称的 Action
	mgr.ActionRegistry().Register("target", []ManifestAction{
		{Name: "otherAction", Version: "1.0"},
	})

	bus := mgr.EventBus()

	// 请求不存在的 Action
	_, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin:  "target",
		Method:        "nonExistAction",
		Payload:       []byte(`{}`),
		ActionVersion: "1.0",
	})
	if !errors.Is(err, ErrActionNotFound) {
		t.Fatalf("期望 ErrActionNotFound，实际: %v", err)
	}
}

// TestCallPlugin_NoVersion_ForwardNormally 不带 action_version → 正常转发（向后兼容）
func TestCallPlugin_NoVersion_ForwardNormally(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	mock := newMockPlugin("target", "1.0.0", "target")
	_ = mgr.Register("target", mock)
	_ = mgr.Start("target")

	bus := mgr.EventBus()

	// 不带 ActionVersion，应直接转发不做版本校验
	resp, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "target",
		Method:       "anyMethod",
		Payload:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("不带版本号应正常转发，实际报错: %v", err)
	}
	if resp == nil || len(resp.Payload) == 0 {
		t.Fatal("响应不应为空")
	}
}

// TestCallPlugin_PluginNotRunning 目标插件未运行 → ErrPluginNotRunning
func TestCallPlugin_PluginNotRunning(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	mock := newMockPlugin("target", "1.0.0", "target")
	_ = mgr.Register("target", mock)
	_ = mgr.Start("target")
	_ = mgr.Stop("target")

	bus := mgr.EventBus()

	_, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "target",
		Method:       "test",
	})
	if !errors.Is(err, ErrPluginNotRunning) {
		t.Fatalf("期望 ErrPluginNotRunning，实际: %v", err)
	}
}

// TestCallPlugin_PluginNotFound 目标插件不存在 → ErrPluginNotFound
func TestCallPlugin_PluginNotFound(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	bus := mgr.EventBus()

	_, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "ghost",
		Method:       "test",
	})
	if !errors.Is(err, ErrPluginNotFound) {
		t.Fatalf("期望 ErrPluginNotFound，实际: %v", err)
	}
}

// TestCallPlugin_VersionCompatible 带 action_version 且版本兼容 → 正常转发
func TestCallPlugin_VersionCompatible(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	mock := newMockPlugin("target", "1.0.0", "target")
	_ = mgr.Register("target", mock)
	_ = mgr.Start("target")

	// 注册 Action 版本 1.2
	mgr.ActionRegistry().Register("target", []ManifestAction{
		{Name: "getData", Version: "1.2"},
	})

	bus := mgr.EventBus()

	// 请求版本 1.0，注册版本 1.2 → major 相同，兼容
	resp, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin:  "target",
		Method:        "getData",
		Payload:       []byte(`{}`),
		ActionVersion: "1.0",
	})
	if err != nil {
		t.Fatalf("版本兼容应正常转发，实际报错: %v", err)
	}
	if resp == nil || len(resp.Payload) == 0 {
		t.Fatal("响应不应为空")
	}
}

// TestPublishEvent_Unchanged PublishEvent 行为不变
func TestPublishEvent_Unchanged(t *testing.T) {
	mgr := NewPluginManager(nil, "")
	mockA := newMockPlugin("pA", "1.0.0", "a")
	mockB := newMockPlugin("pB", "1.0.0", "b")
	_ = mgr.Register("pA", mockA)
	_ = mgr.Register("pB", mockB)
	_ = mgr.Start("pA")
	_ = mgr.Start("pB")

	bus := mgr.EventBus()
	bus.Subscribe("pA", []string{"order.created"})
	bus.Subscribe("pB", []string{"order.created"})

	event := &proto.Event{
		Source:    "host",
		Type:      "order.created",
		Payload:   []byte(`{"id":1}`),
		Timestamp: time.Now().Unix(),
	}
	if err := bus.PublishEvent(context.Background(), event); err != nil {
		t.Fatalf("PublishEvent 失败: %v", err)
	}

	// 两个插件都应收到事件
	eventsA := mockA.getEvents()
	if len(eventsA) != 1 || eventsA[0].Type != "order.created" {
		t.Fatalf("pA 应收到 1 个事件，实际: %d", len(eventsA))
	}
	eventsB := mockB.getEvents()
	if len(eventsB) != 1 || eventsB[0].Type != "order.created" {
		t.Fatalf("pB 应收到 1 个事件，实际: %d", len(eventsB))
	}
}
