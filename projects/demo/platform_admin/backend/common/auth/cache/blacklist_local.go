package cache

import (
	"sync"
	"time"
)

// blacklistEntry 黑名单条目
type blacklistEntry struct {
	expireAt time.Time
}

// LocalBlacklistStore 基于 sync.Map 的本地黑名单存储
type LocalBlacklistStore struct {
	data   sync.Map
	stopCh chan struct{}
}

// NewLocalBlacklistStore 创建本地黑名单存储并启动后台清理 goroutine
func NewLocalBlacklistStore() *LocalBlacklistStore {
	s := &LocalBlacklistStore{
		stopCh: make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Add 将 token 加入黑名单，ttl 后自动过期
func (s *LocalBlacklistStore) Add(token string, ttl time.Duration) error {
	s.data.Store(token, blacklistEntry{
		expireAt: time.Now().Add(ttl),
	})
	return nil
}

// Contains 检查 token 是否在黑名单中（未过期）
func (s *LocalBlacklistStore) Contains(token string) bool {
	val, ok := s.data.Load(token)
	if !ok {
		return false
	}
	entry := val.(blacklistEntry)
	if time.Now().After(entry.expireAt) {
		s.data.Delete(token)
		return false
	}
	return true
}

// Stop 停止后台清理 goroutine
func (s *LocalBlacklistStore) Stop() {
	close(s.stopCh)
}

// cleanup 每 60s 清理过期条目
func (s *LocalBlacklistStore) cleanup() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.data.Range(func(key, value interface{}) bool {
				entry := value.(blacklistEntry)
				if now.After(entry.expireAt) {
					s.data.Delete(key)
				}
				return true
			})
		case <-s.stopCh:
			return
		}
	}
}
