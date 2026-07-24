package spi

import "log"

// SmsSender 短信发送 SPI 接口
type SmsSender interface {
	// SendCode 发送验证码到指定手机号
	SendCode(phone string, code string) error
}

// ConsoleMockSender 控制台打印验证码的 Mock 实现
type ConsoleMockSender struct{}

func (s *ConsoleMockSender) SendCode(phone string, code string) error {
	log.Printf("[SMS-MOCK] 手机号: %s, 验证码: %s", phone, code)
	return nil
}
