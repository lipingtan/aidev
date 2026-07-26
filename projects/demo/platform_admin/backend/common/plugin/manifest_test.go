package plugin

import (
	"encoding/json"
	"testing"
)

func TestManifestV2_FrontendsField(t *testing.T) {
	// 测试含 frontends 字段的 plugin.json
	jsonWithFrontends := `{
        "name": "test-plugin",
        "version": "1.0.0",
        "frontends": [
            {"platform": "admin", "device": "pc", "entry": "admin-pc/bundle.js"},
            {"platform": "user", "device": "h5", "entry": "user-h5/bundle.js"}
        ]
    }`
	var m ManifestV2
	err := json.Unmarshal([]byte(jsonWithFrontends), &m)
	if err != nil {
		t.Fatalf("解析含 frontends 字段的 plugin.json 失败: %v", err)
	}
	if len(m.Frontends) != 2 {
		t.Errorf("期望 2 个 frontend，实际 %d", len(m.Frontends))
	}
	if m.Frontends[0].Platform != "admin" || m.Frontends[0].Device != "pc" {
		t.Errorf("frontends[0] 字段不匹配: %+v", m.Frontends[0])
	}

	// 测试不含 frontends 字段的旧 plugin.json
	jsonWithoutFrontends := `{"name": "old-plugin", "version": "0.0.1"}`
	var m2 ManifestV2
	err2 := json.Unmarshal([]byte(jsonWithoutFrontends), &m2)
	if err2 != nil {
		t.Fatalf("解析不含 frontends 字段失败: %v", err2)
	}
	if len(m2.Frontends) != 0 {
		t.Errorf("期望 frontends 为空，实际 %d", len(m2.Frontends))
	}
}
