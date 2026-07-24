package service

import (
	"errors"
	"regexp"
	"testing"
	"time"
)

// mockSender 测试用 Mock 发送器
type mockSender struct {
	lastPhone string
	lastCode  string
	err       error
}

func (m *mockSender) SendCode(phone string, code string) error {
	m.lastPhone = phone
	m.lastCode = code
	return m.err
}

// newTestSmsService 创建测试用 SmsService（可自定义选项）
func newTestSmsService(opts ...SmsServiceOption) (*SmsService, *mockSender) {
	sender := &mockSender{}
	store := NewMemoryCodeStore()
	svc := NewSmsService(sender, store, opts...)
	return svc, sender
}

// TestSendCode_Generates4DigitCode 验证生成的验证码为 4 位数字
func TestSendCode_Generates4DigitCode(t *testing.T) {
	svc, _ := newTestSmsService()

	code, err := svc.SendCode("13800138000", "tenant1")
	if err != nil {
		t.Fatalf("SendCode 应成功，但返回错误: %v", err)
	}

	matched, _ := regexp.MatchString(`^\d{4}$`, code)
	if !matched {
		t.Errorf("验证码应为 4 位数字，实际: %s", code)
	}
}

// TestVerifyCode_CorrectCode 验证正确验证码校验通过
func TestVerifyCode_CorrectCode(t *testing.T) {
	svc, _ := newTestSmsService()

	code, err := svc.SendCode("13800138000", "tenant1")
	if err != nil {
		t.Fatalf("SendCode 应成功: %v", err)
	}

	err = svc.VerifyCode("13800138000", "tenant1", code)
	if err != nil {
		t.Errorf("正确验证码校验应通过，但返回错误: %v", err)
	}
}

// TestVerifyCode_WrongCode 验证错误验证码校验失败
func TestVerifyCode_WrongCode(t *testing.T) {
	svc, _ := newTestSmsService()

	_, err := svc.SendCode("13800138000", "tenant1")
	if err != nil {
		t.Fatalf("SendCode 应成功: %v", err)
	}

	err = svc.VerifyCode("13800138000", "tenant1", "0000")
	if !errors.Is(err, ErrCodeInvalid) {
		t.Errorf("错误验证码应返回 ErrCodeInvalid，实际: %v", err)
	}
}

// TestVerifyCode_ExpiredCode 验证过期验证码校验失败（通过极短 TTL 模拟）
func TestVerifyCode_ExpiredCode(t *testing.T) {
	svc, _ := newTestSmsService(
		WithCodeTTL(1 * time.Millisecond),
		WithCooldown(1 * time.Millisecond),
	)

	code, err := svc.SendCode("13800138000", "tenant1")
	if err != nil {
		t.Fatalf("SendCode 应成功: %v", err)
	}

	// 等待验证码过期
	time.Sleep(10 * time.Millisecond)

	err = svc.VerifyCode("13800138000", "tenant1", code)
	if !errors.Is(err, ErrCodeExpired) {
		t.Errorf("过期验证码应返回 ErrCodeExpired，实际: %v", err)
	}
}

// TestSendCode_TooFrequent 验证 60 秒内重复发送返回限频错误
func TestSendCode_TooFrequent(t *testing.T) {
	svc, _ := newTestSmsService(
		WithCooldown(60 * time.Second),
	)

	// 第一次发送应成功
	_, err := svc.SendCode("13800138000", "tenant1")
	if err != nil {
		t.Fatalf("第一次 SendCode 应成功: %v", err)
	}

	// 立即再次发送应返回限频错误
	_, err = svc.SendCode("13800138000", "tenant1")
	if err == nil {
		t.Fatal("重复发送应返回错误")
	}

	if !errors.Is(err, ErrTooFrequent) {
		t.Errorf("应返回 ErrTooFrequent，实际: %v", err)
	}

	// 检查返回了剩余秒数
	var tooFreqErr *TooFrequentError
	if errors.As(err, &tooFreqErr) {
		if tooFreqErr.RemainSeconds <= 0 {
			t.Errorf("剩余等待秒数应大于 0，实际: %d", tooFreqErr.RemainSeconds)
		}
	} else {
		t.Errorf("错误应可转为 *TooFrequentError，实际类型: %T", err)
	}
}

// TestVerifyCode_OneTimeUse 验证验证码一次性使用（验证成功后不可重复使用）
func TestVerifyCode_OneTimeUse(t *testing.T) {
	svc, _ := newTestSmsService()

	code, err := svc.SendCode("13800138000", "tenant1")
	if err != nil {
		t.Fatalf("SendCode 应成功: %v", err)
	}

	// 第一次验证成功
	err = svc.VerifyCode("13800138000", "tenant1", code)
	if err != nil {
		t.Fatalf("第一次验证应通过: %v", err)
	}

	// 第二次验证应失败（已被删除）
	err = svc.VerifyCode("13800138000", "tenant1", code)
	if !errors.Is(err, ErrCodeExpired) {
		t.Errorf("验证码已使用后应返回 ErrCodeExpired，实际: %v", err)
	}
}
