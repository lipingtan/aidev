package spi

// NoOpOrganizationProvider OrganizationProvider 的空实现，所有方法返回空切片
type NoOpOrganizationProvider struct{}

func (n *NoOpOrganizationProvider) GetOrgIds(_, _ int64) ([]int64, error) {
	return []int64{}, nil
}

func (n *NoOpOrganizationProvider) GetOrgPath(_, _ int64) ([]int64, error) {
	return []int64{}, nil
}

func (n *NoOpOrganizationProvider) GetSubOrgIds(_, _ int64) ([]int64, error) {
	return []int64{}, nil
}
