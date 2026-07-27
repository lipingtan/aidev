package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitWithTimeout 等待 WaitGroup 完成，超时返回 false。
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

// TestPublish_HandlerReceivesPayload 验证 Publish 后 handler 被异步调用，接收到正确 payload。
func TestPublish_HandlerReceivesPayload(t *testing.T) {
	b := NewBus()

	var wg sync.WaitGroup
	wg.Add(1)

	var received interface{}
	b.Subscribe("test.event", func(payload interface{}) {
		received = payload
		wg.Done()
	})

	b.Publish("test.event", "hello-payload")

	if !waitWithTimeout(&wg, 2*time.Second) {
		t.Fatal("超时：handler 未在预期时间内被调用")
	}
	if received != "hello-payload" {
		t.Fatalf("期望 payload='hello-payload'，实际='%v'", received)
	}
}

// TestPublish_HandlerPanic_DoesNotAffectCaller 验证 handler panic 时调用方不感知，其他 handler 正常触发。
func TestPublish_HandlerPanic_DoesNotAffectCaller(t *testing.T) {
	b := NewBus()

	var wg sync.WaitGroup
	wg.Add(1) // 只等正常的 handler B

	// handler A：panic
	b.Subscribe("test.panic", func(payload interface{}) {
		panic("intentional panic from handler A")
	})

	// handler B：正常执行
	var handlerBCalled bool
	b.Subscribe("test.panic", func(payload interface{}) {
		handlerBCalled = true
		wg.Done()
	})

	// Publish 不应 panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Publish 不应将 panic 传播到调用方，实际收到: %v", r)
			}
		}()
		b.Publish("test.panic", nil)
	}()

	if !waitWithTimeout(&wg, 2*time.Second) {
		t.Fatal("超时：handler B 未在预期时间内被调用")
	}
	if !handlerBCalled {
		t.Fatal("期望 handler B 被调用，实际未被调用")
	}
}

// TestSubscribe_HandlersCalledInOrder 验证两个 handler 按注册顺序 A→B 依次触发。
func TestSubscribe_HandlersCalledInOrder(t *testing.T) {
	b := NewBus()

	var mu sync.Mutex
	order := make([]string, 0, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	b.Subscribe("test.order", func(payload interface{}) {
		mu.Lock()
		order = append(order, "A")
		mu.Unlock()
		wg.Done()
	})
	b.Subscribe("test.order", func(payload interface{}) {
		// 稍微让 A 先落库（goroutine 调度不确定，给 A 一点时间）
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		order = append(order, "B")
		mu.Unlock()
		wg.Done()
	})

	b.Publish("test.order", nil)

	if !waitWithTimeout(&wg, 2*time.Second) {
		t.Fatal("超时：handler 未在预期时间内全部被调用")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 {
		t.Fatalf("期望 2 个 handler 被调用，实际 %d 个", len(order))
	}
	if order[0] != "A" || order[1] != "B" {
		t.Fatalf("期望触发顺序 A→B，实际: %v", order)
	}
}

// TestDefaultBus_IsNotNil 验证全局单例 DefaultBus 可被外部包引用（非 nil，Subscribe+Publish 正常）。
func TestDefaultBus_IsNotNil(t *testing.T) {
	if DefaultBus == nil {
		t.Fatal("DefaultBus 不应为 nil")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	DefaultBus.Subscribe("test.default", func(payload interface{}) {
		wg.Done()
	})
	DefaultBus.Publish("test.default", "singleton")

	if !waitWithTimeout(&wg, 2*time.Second) {
		t.Fatal("超时：DefaultBus handler 未被调用")
	}
}

// TestPublish_NoSubscriber_NoPanic 验证发布无订阅的事件不会 panic。
func TestPublish_NoSubscriber_NoPanic(t *testing.T) {
	b := NewBus()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("无订阅者时 Publish 不应 panic，实际收到: %v", r)
		}
	}()
	b.Publish("no.subscriber", "payload")
}

// TestSubscribe_Concurrent_ThreadSafe 验证并发 Subscribe 不会 data race（配合 -race 使用）。
func TestSubscribe_Concurrent_ThreadSafe(t *testing.T) {
	b := NewBus()
	var wg sync.WaitGroup
	var called int64

	const goroutines = 10
	wg.Add(goroutines)

	// 并发注册 handler
	for i := 0; i < goroutines; i++ {
		go func() {
			b.Subscribe("test.concurrent", func(payload interface{}) {
				atomic.AddInt64(&called, 1)
			})
			wg.Done()
		}()
	}
	wg.Wait()

	// 全部注册完成后 Publish，等待所有 goroutine 执行
	time.Sleep(50 * time.Millisecond)
	b.Publish("test.concurrent", nil)
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&called) != goroutines {
		t.Fatalf("期望 %d 个 handler 被调用，实际 %d 个", goroutines, called)
	}
}
