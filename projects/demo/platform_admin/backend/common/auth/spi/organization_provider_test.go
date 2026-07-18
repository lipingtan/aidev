package spi

import "testing"

// TestNoOpOrganizationProvider_ReturnsEmptySlice 验证 NoOp 实现返回空切片、不报错
func TestNoOpOrganizationProvider_ReturnsEmptySlice(t *testing.T) {
	var p OrganizationProvider = &NoOpOrganizationProvider{}

	ids, err := p.GetOrgIds(1, 1)
	if err != nil {
		t.Fatalf("GetOrgIds: 期望 nil error，得到 %v", err)
	}
	if ids == nil {
		t.Fatal("GetOrgIds: 期望空切片，得到 nil")
	}
	if len(ids) != 0 {
		t.Fatalf("GetOrgIds: 期望长度 0，得到 %d", len(ids))
	}

	path, err := p.GetOrgPath(10, 2)
	if err != nil {
		t.Fatalf("GetOrgPath: 期望 nil error，得到 %v", err)
	}
	if path == nil {
		t.Fatal("GetOrgPath: 期望空切片，得到 nil")
	}
	if len(path) != 0 {
		t.Fatalf("GetOrgPath: 期望长度 0，得到 %d", len(path))
	}

	subIds, err := p.GetSubOrgIds(5, 3)
	if err != nil {
		t.Fatalf("GetSubOrgIds: 期望 nil error，得到 %v", err)
	}
	if subIds == nil {
		t.Fatal("GetSubOrgIds: 期望空切片，得到 nil")
	}
	if len(subIds) != 0 {
		t.Fatalf("GetSubOrgIds: 期望长度 0，得到 %d", len(subIds))
	}
}
