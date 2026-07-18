package spi

import (
	"context"
	"time"
)

// NoOpUserProvider UserProvider 的空实现
type NoOpUserProvider struct{}

func (n *NoOpUserProvider) LoadByUsername(_ string) (*AuthUser, error) {
	return nil, nil
}

// NoOpRoleProvider RoleProvider 的空实现
type NoOpRoleProvider struct{}

func (n *NoOpRoleProvider) GetRolesByUser(_, _ int64) ([]RoleInfo, error) {
	return nil, nil
}

// NoOpCacheAdapter CacheAdapter 的空实现
type NoOpCacheAdapter struct{}

func (n *NoOpCacheAdapter) Get(_ string) (interface{}, bool) {
	return nil, false
}

func (n *NoOpCacheAdapter) Set(_ string, _ interface{}, _ time.Duration) {}

func (n *NoOpCacheAdapter) Delete(_ ...string) {}

func (n *NoOpCacheAdapter) DeleteByPrefix(_ string) {}

// NoOpTokenBlacklistStore TokenBlacklistStore 的空实现
type NoOpTokenBlacklistStore struct{}

func (n *NoOpTokenBlacklistStore) Add(_ string, _ time.Duration) error {
	return nil
}

func (n *NoOpTokenBlacklistStore) Contains(_ string) bool {
	return false
}



// NoOpDataScopeHandler DataScopeHandler 的空实现
type NoOpDataScopeHandler struct{}

func (n *NoOpDataScopeHandler) Handle(_ context.Context, _ string, _ int64) ([]string, error) {
	return nil, nil
}

// NoOpApiDiscoveryStrategy ApiDiscoveryStrategy 的空实现
type NoOpApiDiscoveryStrategy struct{}

func (n *NoOpApiDiscoveryStrategy) Discover(_ interface{}) ([]ApiEndpoint, error) {
	return nil, nil
}

// NoOpEventPublisher EventPublisher 的空实现
type NoOpEventPublisher struct{}

func (n *NoOpEventPublisher) Publish(_ string, _ interface{}) error {
	return nil
}

// NoOpOperationLogger OperationLogger 的空实现
type NoOpOperationLogger struct{}

func (n *NoOpOperationLogger) Log(_ context.Context, _ OperationLogEntry) error {
	return nil
}
