// Package event 提供进程内业务事件总线，异步执行，与插件 EventBus（common/plugin）完全分离。
package event

import "sync"

// Bus 进程内业务事件总线接口。
type Bus interface {
	Subscribe(event string, handler func(payload interface{}))
	Publish(event string, payload interface{})
}

// bus 是 Bus 接口的默认实现，使用 sync.RWMutex 保证订阅注册的线程安全。
type bus struct {
	mu       sync.RWMutex
	handlers map[string][]func(payload interface{})
}

// NewBus 创建一个新的 Bus 实例。
func NewBus() Bus {
	return &bus{
		handlers: make(map[string][]func(payload interface{})),
	}
}

// Subscribe 注册事件处理器，线程安全，多个 handler 按注册顺序存储。
func (b *bus) Subscribe(event string, handler func(payload interface{})) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[event] = append(b.handlers[event], handler)
}

// Publish 异步触发事件，每个 handler 在独立 goroutine 中执行；
// handler panic 会被 recover 捕获，不影响调用方和其他 handler。
func (b *bus) Publish(event string, payload interface{}) {
	b.mu.RLock()
	handlers := make([]func(payload interface{}), len(b.handlers[event]))
	copy(handlers, b.handlers[event])
	b.mu.RUnlock()

	for _, handler := range handlers {
		h := handler
		go func() {
			defer func() { recover() }() //nolint:errcheck
			h(payload)
		}()
	}
}

// DefaultBus 全局默认事件总线单例，供其他包无需依赖注入直接引用。
var DefaultBus = NewBus()

// ---- 事件名常量 ----

const (
	// EventPermissionChanged 权限变更事件（缓存失效触发）
	EventPermissionChanged = "permission.changed"

	// EventAbacPolicyChanged ABAC 策略变更事件（新增，缓存失效触发）
	EventAbacPolicyChanged = "abac.policy.changed"
)

// ---- 事件 payload 类型定义 ----

// PermissionChangedEvent 权限变更事件 payload
type PermissionChangedEvent struct {
	AffectedUsers []AffectedUser // 受影响的用户列表
	Source        string         // 变更来源标识（"assign_resources"/"assign_apis"/"replace_roles" 等）
}

// AffectedUser 受影响的用户
type AffectedUser struct {
	UserID   int64
	TenantID int64
}

// ApprovalCompletedEvent 审批通过事件。
type ApprovalCompletedEvent struct {
	ApprovalID int64
	BizType    string
	BizID      string
	TenantID   int64
}

// ApprovalRejectedEvent 审批驳回事件。
type ApprovalRejectedEvent struct {
	ApprovalID int64
	BizType    string
	BizID      string
	TenantID   int64
}

// ApprovalCancelledEvent 审批撤销事件。
type ApprovalCancelledEvent struct {
	ApprovalID int64
	BizType    string
	BizID      string
	TenantID   int64
}

// AbacPolicyChangedEvent ABAC 策略变更事件（缓存失效触发）
type AbacPolicyChangedEvent struct {
	TenantID     int64  // 变更的策略所属租户 ID（0=平台级）
	ResourceType string // 变更的资源类型
}
