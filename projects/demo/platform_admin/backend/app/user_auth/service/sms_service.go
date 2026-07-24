package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go-admin/app/user_auth/spi"
)

var (
	// ErrTooFrequent 发送过于频繁
	ErrTooFrequent = errors.New("发送过于频繁")
	// ErrCodeInvalid 验证码错误
	ErrCodeInvalid = errors.New("验证码错误")
	// ErrCodeExpired 验证码已过期或不存在
	ErrCodeExpired = errors.New("验证码已过期或不存在")
)

// ──────────────────────────────────────────────────────────────────────────────
// CodeStore 验证码存储接口（便于未来替换为 Redis）
// ──────────────────────────────────────────────────────────────────────────────

// CodeStore 验证码存储接口
type CodeStore interface {
	// Set 设置 key-value 并指定过期时间
	Set(key string, value string, ttl time.Duration) error
	// Get 获取 key 对应的 value，不存在返回空字符串和 error
	Get(key string) (string, error)
	// Delete 删除指定 key
	Delete(key string) error
	// Exists 检查 key 是否存在
	Exists(key string) bool
}

// ──────────────────────────────────────────────────────────────────────────────
// MemoryCodeStore 内存实现（sync.Map）
// 注意：仅适用于单实例部署
// ──────────────────────────────────────────────────────────────────────────────

type memoryEntry struct {
	value     string
	expireAt  time.Time
}

// MemoryCodeStore 基于内存的 CodeStore 实现
type MemoryCodeStore struct {
	data sync.Map
}

// NewMemoryCodeStore 创建内存存储实例
func NewMemoryCodeStore() *MemoryCodeStore {
	return &MemoryCodeStore{}
}

func (m *MemoryCodeStore) Set(key string, value string, ttl time.Duration) error {
	m.data.Store(key, &memoryEntry{
		value:    value,
		expireAt: time.Now().Add(ttl),
	})
	return nil
}

func (m *MemoryCodeStore) Get(key string) (string, error) {
	val, ok := m.data.Load(key)
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	entry := val.(*memoryEntry)
	if time.Now().After(entry.expireAt) {
		m.data.Delete(key)
		return "", fmt.Errorf("key expired: %s", key)
	}
	return entry.value, nil
}

func (m *MemoryCodeStore) Delete(key string) error {
	m.data.Delete(key)
	return nil
}

func (m *MemoryCodeStore) Exists(key string) bool {
	val, ok := m.data.Load(key)
	if !ok {
		return false
	}
	entry := val.(*memoryEntry)
	if time.Now().After(entry.expireAt) {
		m.data.Delete(key)
		return false
	}
	return true
}

// ──────────────────────────────────────────────────────────────────────────────
// SmsService 短信验证码服务
// ──────────────────────────────────────────────────────────────────────────────

// SmsService 短信验证码服务
type SmsService struct {
	sender   spi.SmsSender
	store    CodeStore
	codeTTL  time.Duration // 验证码有效期
	codeLen  int           // 验证码长度
	cooldown time.Duration // 发送冷却时间
}

// SmsServiceOption 配置选项
type SmsServiceOption func(*SmsService)

// WithCodeTTL 设置验证码有效期
func WithCodeTTL(ttl time.Duration) SmsServiceOption {
	return func(s *SmsService) {
		s.codeTTL = ttl
	}
}

// WithCodeLen 设置验证码长度
func WithCodeLen(length int) SmsServiceOption {
	return func(s *SmsService) {
		s.codeLen = length
	}
}

// WithCooldown 设置发送冷却时间
func WithCooldown(cd time.Duration) SmsServiceOption {
	return func(s *SmsService) {
		s.cooldown = cd
	}
}

// NewSmsService 创建短信验证码服务
func NewSmsService(sender spi.SmsSender, store CodeStore, opts ...SmsServiceOption) *SmsService {
	svc := &SmsService{
		sender:   sender,
		store:    store,
		codeTTL:  5 * time.Minute,
		codeLen:  4,
		cooldown: 60 * time.Second,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// TooFrequentError 限频错误，包含剩余等待秒数
type TooFrequentError struct {
	RemainSeconds int
}

func (e *TooFrequentError) Error() string {
	return fmt.Sprintf("发送过于频繁，请 %d 秒后重试", e.RemainSeconds)
}

func (e *TooFrequentError) Is(target error) bool {
	return target == ErrTooFrequent
}

// SendCode 发送验证码
// 1. 检查冷却时间（限频）
// 2. 生成随机数字验证码
// 3. 存储验证码（key=code:{phone}:{tenantCode}）
// 4. 设置冷却标记（key=cd:{phone}:{tenantCode}）
// 5. 调用 SmsSender 发送
func (s *SmsService) SendCode(phone string, tenantCode string) (string, error) {
	cdKey := fmt.Sprintf("cd:%s:%s", phone, tenantCode)

	// 检查冷却时间
	if s.store.Exists(cdKey) {
		cdVal, err := s.store.Get(cdKey)
		if err == nil {
			// 计算剩余等待秒数
			sendTime, parseErr := time.Parse(time.RFC3339Nano, cdVal)
			if parseErr == nil {
				elapsed := time.Since(sendTime)
				remain := s.cooldown - elapsed
				if remain > 0 {
					return "", &TooFrequentError{RemainSeconds: int(remain.Seconds()) + 1}
				}
			}
		}
		// 如果解析失败但 key 存在，仍视为限频中
		return "", &TooFrequentError{RemainSeconds: int(s.cooldown.Seconds())}
	}

	// 生成验证码
	code, err := s.generateCode()
	if err != nil {
		return "", fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存储验证码
	codeKey := fmt.Sprintf("code:%s:%s", phone, tenantCode)
	if err := s.store.Set(codeKey, code, s.codeTTL); err != nil {
		return "", fmt.Errorf("存储验证码失败: %w", err)
	}

	// 设置冷却标记
	if err := s.store.Set(cdKey, time.Now().Format(time.RFC3339Nano), s.cooldown); err != nil {
		return "", fmt.Errorf("设置冷却标记失败: %w", err)
	}

	// 发送验证码
	if err := s.sender.SendCode(phone, code); err != nil {
		return "", fmt.Errorf("发送验证码失败: %w", err)
	}

	return code, nil
}

// VerifyCode 校验验证码
// 1. 从 store 获取存储的验证码
// 2. 比对
// 3. 成功后删除（一次性使用）
func (s *SmsService) VerifyCode(phone string, tenantCode string, inputCode string) error {
	codeKey := fmt.Sprintf("code:%s:%s", phone, tenantCode)

	storedCode, err := s.store.Get(codeKey)
	if err != nil {
		return ErrCodeExpired
	}

	if storedCode != inputCode {
		return ErrCodeInvalid
	}

	// 验证成功，删除验证码（一次性使用）
	_ = s.store.Delete(codeKey)
	return nil
}

// generateCode 生成指定长度的随机数字验证码
func (s *SmsService) generateCode() (string, error) {
	code := ""
	for i := 0; i < s.codeLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
