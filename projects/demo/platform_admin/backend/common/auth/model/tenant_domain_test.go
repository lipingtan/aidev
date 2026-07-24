package model

import (
	"reflect"
	"strings"
	"testing"
)

// TestTenantDomainTableName 验证表名正确
func TestTenantDomainTableName(t *testing.T) {
	td := TenantDomain{}
	if got := td.TableName(); got != "tenant_domain" {
		t.Errorf("TableName() = %q, want %q", got, "tenant_domain")
	}
}

// TestTenantDomainJSONTags 验证 ID、TenantID 含 json:",string" tag
func TestTenantDomainJSONTags(t *testing.T) {
	typ := reflect.TypeOf(TenantDomain{})

	idField, ok := typ.FieldByName("ID")
	if !ok {
		t.Fatal("TenantDomain 缺少 ID 字段")
	}
	if tag := idField.Tag.Get("json"); !strings.Contains(tag, "string") {
		t.Errorf("ID json tag = %q, 应包含 ',string'", tag)
	}

	tenantIDField, ok := typ.FieldByName("TenantID")
	if !ok {
		t.Fatal("TenantDomain 缺少 TenantID 字段")
	}
	if tag := tenantIDField.Tag.Get("json"); !strings.Contains(tag, "string") {
		t.Errorf("TenantID json tag = %q, 应包含 ',string'", tag)
	}
}

// TestTenantDomainBeforeCreate_AutoID 验证 BeforeCreate 自动生成雪花 ID
func TestTenantDomainBeforeCreate_AutoID(t *testing.T) {
	td := &TenantDomain{} // ID = 0
	if err := td.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate 返回 error: %v", err)
	}
	if td.ID == 0 {
		t.Error("BeforeCreate 未设置 ID，期望非零雪花 ID")
	}
}

// TestTenantDomainBeforeCreate_KeepExistingID 已有 ID 时不覆盖
func TestTenantDomainBeforeCreate_KeepExistingID(t *testing.T) {
	const existingID int64 = 12345
	td := &TenantDomain{ID: existingID}
	if err := td.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate 返回 error: %v", err)
	}
	if td.ID != existingID {
		t.Errorf("BeforeCreate 修改了已有 ID: got %d, want %d", td.ID, existingID)
	}
}
