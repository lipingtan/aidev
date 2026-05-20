package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"game-server/plugin-sdk/proto"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// Mock 插件实现
// =============================================================================

// mockPlugin 模拟插件，实现 proto.PluginService 接口
type mockPlugin struct {
	name        string
	version     string
	routePrefix string
	events      []proto.Event // 收到的事件记录
	mu          sync.Mutex
}

func newMockPlugin(name, version, routePrefix string) *mockPlugin {
	return &mockPlugin{
		name:        name,
		version:     version,
		routePrefix: routePrefix,
	}
}

func (p *mockPlugin) Register(ctx context.Context) (*proto.PluginInfo, error) {
	return &proto.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: p.name + " 测试插件",
		RoutePrefix: p.routePrefix,
		Menus: []proto.MenuItem{
			{Title: p.name, Icon: "ep/box", Path: "/plugin/" + p.name, Sort: 1},
		},
		Perms: []proto.Permission{
			{Key: p.name + ":list", Title: p.name + "列表"},
		},
	}, nil
}

func (p *mockPlugin) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	body := fmt.Sprintf(`{"plugin":"%s","method":"%s","path":"%s","userId":%d,"tenantId":%d}`,
		p.name, req.Method, req.Path, req.UserID, req.TenantID)
	return &proto.HttpResponse{
		StatusCode: int32(http.StatusOK),
		Headers:    map[string]string{"Content-Type": "application/json", "X-Plugin": p.name},
		Body:       []byte(body),
	}, nil
}

func (p *mockPlugin) Healthcheck(ctx context.Context) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{Healthy: true, Message: "ok"}, nil
}

func (p *mockPlugin) OnEvent(ctx context.Context, event *proto.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, *event)
	return nil
}

func (p *mockPlugin) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	resp := fmt.Sprintf(`{"from":"%s","method":"%s"}`, p.name, req.Method)
	return &proto.CallPluginResponse{Payload: []byte(resp)}, nil
}

func (p *mockPlugin) getEvents() []proto.Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := make([]proto.Event, len(p.events))
	copy(cp, p.events)
	return cp
}

// =============================================================================
// 测试用例
// =============================================================================

// TestPluginLifecycle 测试插件完整生命周期：注册 → 启动 → 健康检查 → 停止
func TestPluginLifecycle(t *testing.T) {
	mgr := NewPluginManager()
	mock := newMockPlugin("game", "1.0.0", "game")

	// 注册
	if err := mgr.Register("game", mock); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	// 重复注册应报错
	if err := mgr.Register("game", mock); err == nil {
		t.Fatal("重复注册应返回错误")
	}

	// 启动前状态应为 Stopped
	inst, ok := mgr.GetPlugin("game")
	if !ok {
		t.Fatal("GetPlugin 应返回已注册的插件")
	}
	if inst.Status != StatusStopped {
		t.Fatalf("启动前状态应为 Stopped，实际: %d", inst.Status)
	}

	// 启动
	if err := mgr.Start("game"); err != nil {
		t.Fatalf("启动失败: %v", err)
	}

	// 启动后状态应为 Running
	inst, _ = mgr.GetPlugin("game")
	if inst.Status != StatusRunning {
		t.Fatalf("启动后状态应为 Running，实际: %d", inst.Status)
	}
	if inst.Info == nil || inst.Info.Name != "game" {
		t.Fatal("启动后 Info 应被填充")
	}

	// 重复启动应报错
	if err := mgr.Start("game"); err == nil {
		t.Fatal("重复启动应返回错误")
	}

	// 健康检查
	resp, err := mgr.Healthcheck("game")
	if err != nil {
		t.Fatalf("健康检查失败: %v", err)
	}
	if !resp.Healthy {
		t.Fatal("健康检查应返回 healthy=true")
	}

	// 停止
	if err := mgr.Stop("game"); err != nil {
		t.Fatalf("停止失败: %v", err)
	}
	inst, _ = mgr.GetPlugin("game")
	if inst.Status != StatusStopped {
		t.Fatalf("停止后状态应为 Stopped，实际: %d", inst.Status)
	}

	// 停止后健康检查应返回 unhealthy
	resp, err = mgr.Healthcheck("game")
	if err != nil {
		t.Fatalf("停止后健康检查不应报错: %v", err)
	}
	if resp.Healthy {
		t.Fatal("停止后健康检查应返回 healthy=false")
	}
}

// TestRegistry 测试路由/菜单/权限注册和注销
func TestRegistry(t *testing.T) {
	mgr := NewPluginManager()
	mock := newMockPlugin("shop", "1.0.0", "shop")

	_ = mgr.Register("shop", mock)
	_ = mgr.Start("shop")

	// 验证路由注册
	reg := mgr.Registry()
	name, found := reg.FindPluginByRoute("shop/orders")
	if !found || name != "shop" {
		t.Fatalf("路由查找失败: found=%v, name=%s", found, name)
	}

	// 验证菜单注册
	menus := reg.GetAllMenus()
	if _, ok := menus["shop"]; !ok {
		t.Fatal("菜单应包含 shop 插件")
	}

	// 验证权限注册
	perms := reg.GetAllPermissions()
	if len(perms) == 0 {
		t.Fatal("权限列表不应为空")
	}

	// 停止后应注销
	_ = mgr.Stop("shop")
	_, found = reg.FindPluginByRoute("shop/orders")
	if found {
		t.Fatal("停止后路由应被注销")
	}
	menus = reg.GetAllMenus()
	if _, ok := menus["shop"]; ok {
		t.Fatal("停止后菜单应被注销")
	}
}

// TestEventBus 测试事件广播和插件间调用
func TestEventBus(t *testing.T) {
	mgr := NewPluginManager()
	pluginA := newMockPlugin("pluginA", "1.0.0", "a")
	pluginB := newMockPlugin("pluginB", "1.0.0", "b")

	_ = mgr.Register("pluginA", pluginA)
	_ = mgr.Register("pluginB", pluginB)
	_ = mgr.Start("pluginA")
	_ = mgr.Start("pluginB")

	// 订阅事件
	bus := mgr.EventBus()
	bus.Subscribe("pluginA", []string{"order.created", "user.login"})
	bus.Subscribe("pluginB", []string{"order.created"})

	// 发布事件
	event := &proto.Event{
		Source:    "host",
		Type:      "order.created",
		Payload:   []byte(`{"orderId":123}`),
		Timestamp: time.Now().Unix(),
	}
	if err := bus.PublishEvent(context.Background(), event); err != nil {
		t.Fatalf("发布事件失败: %v", err)
	}

	// 验证两个插件都收到了事件
	eventsA := pluginA.getEvents()
	if len(eventsA) != 1 || eventsA[0].Type != "order.created" {
		t.Fatalf("pluginA 应收到 1 个 order.created 事件，实际: %d", len(eventsA))
	}
	eventsB := pluginB.getEvents()
	if len(eventsB) != 1 || eventsB[0].Type != "order.created" {
		t.Fatalf("pluginB 应收到 1 个 order.created 事件，实际: %d", len(eventsB))
	}

	// 发布 user.login 事件，只有 pluginA 订阅
	event2 := &proto.Event{Source: "host", Type: "user.login", Payload: []byte(`{}`), Timestamp: time.Now().Unix()}
	_ = bus.PublishEvent(context.Background(), event2)

	eventsA = pluginA.getEvents()
	if len(eventsA) != 2 {
		t.Fatalf("pluginA 应收到 2 个事件，实际: %d", len(eventsA))
	}
	eventsB = pluginB.getEvents()
	if len(eventsB) != 1 {
		t.Fatalf("pluginB 应仍只有 1 个事件，实际: %d", len(eventsB))
	}

	// 测试 CallPlugin：A 调用 B
	callResp, err := bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "pluginB",
		Method:       "getData",
		Payload:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("CallPlugin 失败: %v", err)
	}
	if !strings.Contains(string(callResp.Payload), "pluginB") {
		t.Fatalf("CallPlugin 响应应来自 pluginB，实际: %s", string(callResp.Payload))
	}

	// 调用不存在的插件应报错
	_, err = bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "nonexist",
		Method:       "test",
	})
	if err == nil {
		t.Fatal("调用不存在的插件应返回错误")
	}

	// 停止 pluginB 后调用应报错
	_ = mgr.Stop("pluginB")
	_, err = bus.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "pluginB",
		Method:       "test",
	})
	if err == nil {
		t.Fatal("调用已停止的插件应返回错误")
	}
}

// TestHostService 测试 HostService 接口
func TestHostService(t *testing.T) {
	mgr := NewPluginManager()
	mock := newMockPlugin("svc", "1.0.0", "svc")
	_ = mgr.Register("svc", mock)
	_ = mgr.Start("svc")

	bus := mgr.EventBus()
	bus.Subscribe("svc", []string{"test.event"})

	hostSvc := mgr.HostService()

	// PublishEvent
	err := hostSvc.PublishEvent(context.Background(), &proto.Event{
		Source: "external", Type: "test.event", Payload: []byte(`{"x":1}`),
	})
	if err != nil {
		t.Fatalf("HostService.PublishEvent 失败: %v", err)
	}
	events := mock.getEvents()
	if len(events) != 1 {
		t.Fatalf("应收到 1 个事件，实际: %d", len(events))
	}

	// CallPlugin
	resp, err := hostSvc.CallPlugin(context.Background(), &proto.CallPluginRequest{
		TargetPlugin: "svc", Method: "hello",
	})
	if err != nil {
		t.Fatalf("HostService.CallPlugin 失败: %v", err)
	}
	if !strings.Contains(string(resp.Payload), "svc") {
		t.Fatalf("响应应包含 svc，实际: %s", string(resp.Payload))
	}
}

// TestPluginProxy 测试 HTTP 代理转发
func TestPluginProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mgr := NewPluginManager()
	mock := newMockPlugin("demo", "1.0.0", "demo")
	_ = mgr.Register("demo", mock)
	_ = mgr.Start("demo")

	// 创建 gin 路由并注册代理
	r := gin.New()
	proxy := NewPluginProxy(mgr)
	v1 := r.Group("/api/v1")
	proxy.RegisterRoutes(v1)

	// 测试正常代理请求
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugin/demo/orders?page=1", nil)
	w := httptest.NewRecorder()

	// 模拟中间件注入的上下文
	r.Use(func(c *gin.Context) {
		c.Set("userId", int64(42))
		c.Set("tenantId", int64(100))
		c.Set("roles", []string{"admin"})
		c.Set("permissions", []string{"demo:list"})
		c.Next()
	})

	// 重新注册（中间件在前）
	r2 := gin.New()
	r2.Use(func(c *gin.Context) {
		c.Set("userId", int64(42))
		c.Set("tenantId", int64(100))
		c.Set("roles", []string{"admin"})
		c.Set("permissions", []string{"demo:list"})
		c.Next()
	})
	proxy2 := NewPluginProxy(mgr)
	v1_2 := r2.Group("/api/v1")
	proxy2.RegisterRoutes(v1_2)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugin/demo/orders?page=1", nil)
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("代理请求应返回 200，实际: %d, body: %s", w.Code, w.Body.String())
	}

	// 验证响应内容
	var respBody map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}
	if respBody["plugin"] != "demo" {
		t.Fatalf("响应 plugin 应为 demo，实际: %v", respBody["plugin"])
	}
	if respBody["path"] != "/orders" {
		t.Fatalf("响应 path 应为 /orders，实际: %v", respBody["path"])
	}
	if int64(respBody["userId"].(float64)) != 42 {
		t.Fatalf("响应 userId 应为 42，实际: %v", respBody["userId"])
	}
	if int64(respBody["tenantId"].(float64)) != 100 {
		t.Fatalf("响应 tenantId 应为 100，实际: %v", respBody["tenantId"])
	}

	// 验证自定义响应头
	if w.Header().Get("X-Plugin") != "demo" {
		t.Fatalf("响应头 X-Plugin 应为 demo，实际: %s", w.Header().Get("X-Plugin"))
	}

	// 测试插件不存在时返回 503
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugin/nonexist/test", nil)
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("不存在的插件应返回 503，实际: %d", w.Code)
	}

	// 测试插件停止后返回 503
	_ = mgr.Stop("demo")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugin/demo/test", nil)
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("已停止的插件应返回 503，实际: %d", w.Code)
	}
}

// TestStopAll 测试批量停止
func TestStopAll(t *testing.T) {
	mgr := NewPluginManager()
	_ = mgr.Register("p1", newMockPlugin("p1", "1.0.0", "p1"))
	_ = mgr.Register("p2", newMockPlugin("p2", "1.0.0", "p2"))
	_ = mgr.Start("p1")
	_ = mgr.Start("p2")

	mgr.StopAll()

	for _, name := range []string{"p1", "p2"} {
		inst, _ := mgr.GetPlugin(name)
		if inst.Status != StatusStopped {
			t.Fatalf("StopAll 后 %s 应为 Stopped", name)
		}
	}

	// Registry 应清空
	reg := mgr.Registry()
	if _, found := reg.FindPluginByRoute("p1"); found {
		t.Fatal("StopAll 后路由应被清空")
	}
}

// TestListPlugins 测试列出所有插件
func TestListPlugins(t *testing.T) {
	mgr := NewPluginManager()
	_ = mgr.Register("a", newMockPlugin("a", "1.0.0", "a"))
	_ = mgr.Register("b", newMockPlugin("b", "2.0.0", "b"))

	list := mgr.ListPlugins()
	if len(list) != 2 {
		t.Fatalf("应有 2 个插件，实际: %d", len(list))
	}
}

// TestEventBusUnsubscribeOnStop 测试停止插件后事件不再广播到该插件
func TestEventBusUnsubscribeOnStop(t *testing.T) {
	mgr := NewPluginManager()
	mock := newMockPlugin("sub", "1.0.0", "sub")
	_ = mgr.Register("sub", mock)
	_ = mgr.Start("sub")

	bus := mgr.EventBus()
	bus.Subscribe("sub", []string{"notify"})

	// 停止后发布事件
	_ = mgr.Stop("sub")
	_ = bus.PublishEvent(context.Background(), &proto.Event{Type: "notify", Payload: []byte(`{}`)})

	events := mock.getEvents()
	if len(events) != 0 {
		t.Fatalf("停止后不应收到事件，实际收到: %d", len(events))
	}
}
