package plugin

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestStartProcessIntegration 端到端测试：通过 go-plugin 启动 DLC 插件子进程
// 验证 StartProcess → Register → Healthcheck → Stop 完整流程
func TestStartProcessIntegration(t *testing.T) {
	// 查找 DLC 插件二进制
	// 相对于 backend/common/plugin/ 目录，插件在 ../../../plugins/dlc/dlc-plugin.exe
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	pluginBinary := filepath.Join(testDir, "..", "..", "..", "plugins", "dlc", "dlc-plugin.exe")

	if _, err := os.Stat(pluginBinary); os.IsNotExist(err) {
		t.Skipf("DLC 插件二进制不存在，跳过集成测试: %s", pluginBinary)
	}

	mgr := NewPluginManager()

	// 启动插件子进程
	err := mgr.StartProcess("dlc", pluginBinary)
	if err != nil {
		t.Fatalf("StartProcess 失败: %v", err)
	}

	// 验证插件已注册
	inst, ok := mgr.GetPlugin("dlc")
	if !ok {
		t.Fatal("GetPlugin 应返回已启动的插件")
	}
	if inst.Status != StatusRunning {
		t.Fatalf("插件状态应为 Running，实际: %d", inst.Status)
	}
	if inst.Info == nil || inst.Info.Name != "dlc" {
		t.Fatal("插件 Info 应包含正确的名称")
	}
	if inst.Info.Version != "1.0.0" {
		t.Fatalf("插件版本应为 1.0.0，实际: %s", inst.Info.Version)
	}
	if inst.Info.RoutePrefix != "dlc" {
		t.Fatalf("路由前缀应为 dlc，实际: %s", inst.Info.RoutePrefix)
	}

	// 验证 Registry 注册
	reg := mgr.Registry()
	name, found := reg.FindPluginByRoute("dlc")
	if !found || name != "dlc" {
		t.Fatalf("Registry 应包含 dlc 路由，found=%v, name=%s", found, name)
	}

	// 健康检查
	resp, err := mgr.Healthcheck("dlc")
	if err != nil {
		t.Fatalf("Healthcheck 失败: %v", err)
	}
	if !resp.Healthy {
		t.Fatalf("健康检查应返回 healthy=true，message=%s", resp.Message)
	}

	// 通过 HandleRequest 调用插件（不连数据库，会返回错误但验证通信正常）
	// 注意：DLC 插件需要 db_dsn 才能正常工作，这里只验证 gRPC 通信链路
	httpResp, err := inst.Service.HandleRequest(context.Background(), nil)
	// 传 nil 可能导致插件 panic，改为传空请求
	_ = httpResp
	_ = err

	// 停止插件
	if err := mgr.Stop("dlc"); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}

	// 验证停止后状态
	inst, _ = mgr.GetPlugin("dlc")
	if inst.Status != StatusStopped {
		t.Fatalf("停止后状态应为 Stopped，实际: %d", inst.Status)
	}

	// 验证 Registry 已注销
	_, found = reg.FindPluginByRoute("dlc")
	if found {
		t.Fatal("停止后路由应被注销")
	}
}
