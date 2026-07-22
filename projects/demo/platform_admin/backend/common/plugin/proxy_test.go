package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"platform-admin/plugin-sdk/proto"
)

// mockPluginService 测试用 mock 实现
type mockPluginService struct{}

func (m *mockPluginService) Register(_ context.Context) (*proto.PluginInfo, error) {
	return &proto.PluginInfo{Name: "test-plugin"}, nil
}

func (m *mockPluginService) HandleRequest(_ context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
	return &proto.HttpResponse{
		StatusCode: 200,
		Headers:    map[string]string{"X-Plugin": "test-plugin"},
		Body:       []byte(`{"ok":true}`),
	}, nil
}

func (m *mockPluginService) Healthcheck(_ context.Context) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{Healthy: true}, nil
}

func (m *mockPluginService) OnEvent(_ context.Context, _ *proto.Event) error {
	return nil
}

func (m *mockPluginService) CallPlugin(_ context.Context, _ *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
	return &proto.CallPluginResponse{}, nil
}

// setupTestProxy 构建测试用 PluginProxy 和 gin 引擎
func setupTestProxy() (*PluginProxy, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	// 创建 PluginManager（传 nil db，测试不涉及数据库）
	mgr := NewPluginManager(nil, "")

	// 注册一个运行中的插件
	_ = mgr.Register("running-plugin", &mockPluginService{})
	inst, _ := mgr.GetPlugin("running-plugin")
	inst.Status = StatusRunning

	// 注册一个已停止的插件
	_ = mgr.Register("stopped-plugin", &mockPluginService{})
	// Register 默认 StatusStopped，无需额外设置

	proxy := NewPluginProxy(mgr, "test-dsn")

	r := gin.New()
	api := r.Group("/api/v1")
	proxy.RegisterRoutes(api)

	return proxy, r
}

// TestProxy_NotFound 请求不存在的插件应返回 404 + code=40400
func TestProxy_NotFound(t *testing.T) {
	_, r := setupTestProxy()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/plugin/nonexistent/hello", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("期望状态码 404，实际 %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	code, ok := body["code"].(float64)
	if !ok || int(code) != 40400 {
		t.Fatalf("期望 code=40400，实际 %v", body["code"])
	}

	msg, ok := body["msg"].(string)
	if !ok || msg != "插件不存在" {
		t.Fatalf("期望 msg='插件不存在'，实际 %v", body["msg"])
	}
}

// TestProxy_Stopped 请求已停止的插件应返回 503 + code=50301 + data
func TestProxy_Stopped(t *testing.T) {
	_, r := setupTestProxy()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/plugin/stopped-plugin/action", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("期望状态码 503，实际 %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}

	code, ok := body["code"].(float64)
	if !ok || int(code) != 50301 {
		t.Fatalf("期望 code=50301，实际 %v", body["code"])
	}

	msg, ok := body["msg"].(string)
	if !ok || msg != "插件维护中，请稍后再试" {
		t.Fatalf("期望 msg='插件维护中，请稍后再试'，实际 %v", body["msg"])
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("期望 data 为对象，实际 %v", body["data"])
	}
	if data["plugin"] != "stopped-plugin" {
		t.Fatalf("期望 data.plugin='stopped-plugin'，实际 %v", data["plugin"])
	}
	if data["status"] != "stopped" {
		t.Fatalf("期望 data.status='stopped'，实际 %v", data["status"])
	}
}

// TestProxy_Running 请求运行中的插件应正常代理转发
func TestProxy_Running(t *testing.T) {
	_, r := setupTestProxy()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/plugin/running-plugin/test-action", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", w.Code)
	}

	// 验证插件响应头被转发
	if w.Header().Get("X-Plugin") != "test-plugin" {
		t.Fatalf("期望响应头 X-Plugin='test-plugin'，实际 '%s'", w.Header().Get("X-Plugin"))
	}

	// 验证响应体被转发
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应 JSON 失败: %v", err)
	}
	if body["ok"] != true {
		t.Fatalf("期望 body.ok=true，实际 %v", body["ok"])
	}
}
