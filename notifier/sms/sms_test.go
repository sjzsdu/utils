package sms

import (
	"testing"
)

// MockMessageItem 用于测试的模拟消息项
type MockMessageItem struct {
	mockTitle   string
	mockURL     string
	mockContent string
}

func (m *MockMessageItem) Title() string {
	return m.mockTitle
}

func (m *MockMessageItem) URL() string {
	return m.mockURL
}

func (m *MockMessageItem) Content() string {
	return m.mockContent
}

// 测试创建短信通知器
func TestNewSMSNotifier(t *testing.T) {
	// 创建禁用的配置
	disabledConfig := &SMSNotifierConfig{
		Enabled:      false,
		Provider:     "alicloud",
		AccessKey:    "test_access_key",
		SecretKey:    "test_secret_key",
		Signature:    "测试签名",
		TemplateID:   "SMS_12345678",
		PhoneNumbers: []string{"13800138000"},
	}

	// 测试创建禁用的通知器
	disabledNotifier, err := NewNotifier(disabledConfig)
	if err != nil {
		t.Fatalf("创建禁用的短信通知器失败: %v", err)
	}
	if disabledNotifier.IsEnabled() {
		t.Error("禁用的通知器应该返回false")
	}

	// 创建启用的配置
	enabledConfig := &SMSNotifierConfig{
		Enabled:      true,
		Provider:     "alicloud",
		AccessKey:    "test_access_key",
		SecretKey:    "test_secret_key",
		Signature:    "测试签名",
		TemplateID:   "SMS_12345678",
		PhoneNumbers: []string{"13800138000"},
	}

	// 测试创建启用的通知器
	enabledNotifier, err := NewNotifier(enabledConfig)
	if err != nil {
		t.Fatalf("创建启用的短信通知器失败: %v", err)
	}
	if !enabledNotifier.IsEnabled() {
		t.Error("启用的通知器应该返回true")
	}

	// 验证通知器名称
	if enabledNotifier.Name() != "sms" {
		t.Errorf("通知器名称不匹配，期望'sms'，实际得到'%s'", enabledNotifier.Name())
	}
}
