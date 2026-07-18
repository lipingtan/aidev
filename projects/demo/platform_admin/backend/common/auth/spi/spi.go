// Package spi 定义 auth-rbac 模块的 SPI 接口，支持构造函数注入
package spi

import (
	"context"
	"time"
)

// AuthUser 认证用户信息
type AuthUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Status   int    `json:"status"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID       int64  `json:"id"`
	RoleCode string `json:"roleCode"`
	RoleName string `json:"roleName"`
	RoleType string `json:"roleType"`
}

// ApiEndpoint API 端点信息
type ApiEndpoint struct {
	Method         string `json:"method"`
	Path           string `json:"path"`
	Name           string `json:"name"`
	PermissionCode string `json:"permissionCode"`
}

// OperationLogEntry 操作日志条目
type OperationLogEntry struct {
	Module     string      `json:"module"`
	Action     string      `json:"action"`
	TargetType string      `json:"targetType"`
	TargetID   int64       `json:"targetId"`
	Summary    string      `json:"summary"`
	OldValue   interface{} `json:"oldValue"`
	NewValue   interface{} `json:"newValue"`
}

// UserProvider 用户数据提供者
type UserProvider interface {
	LoadByUsername(username string) (*AuthUser, error)
}

// RoleProvider 角色数据提供者
type RoleProvider interface {
	GetRolesByUser(userID, tenantID int64) ([]RoleInfo, error)
}

// CacheAdapter 缓存适配器
type CacheAdapter interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(keys ...string)
	DeleteByPrefix(prefix string)
}

// TokenBlacklistStore Token 黑名单存储
type TokenBlacklistStore interface {
	Add(token string, ttl time.Duration) error
	Contains(token string) bool
}

// OrganizationProvider 组织架构数据提供者
type OrganizationProvider interface {
	GetOrgIds(userID, tenantID int64) ([]int64, error)
	GetOrgPath(orgID, tenantID int64) ([]int64, error)
	GetSubOrgIds(orgID, tenantID int64) ([]int64, error)
}

// DataScopeHandler 数据权限处理器
type DataScopeHandler interface {
	Handle(ctx context.Context, dimension string, userID int64) ([]string, error)
}

// ApiDiscoveryStrategy API 自动发现策略
type ApiDiscoveryStrategy interface {
	Discover(engine interface{}) ([]ApiEndpoint, error)
}

// EventPublisher 事件发布器
type EventPublisher interface {
	Publish(event string, payload interface{}) error
}

// OperationLogger 操作日志记录器
type OperationLogger interface {
	Log(ctx context.Context, entry OperationLogEntry) error
}
