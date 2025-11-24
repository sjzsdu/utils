package email

import (
	"testing"
)

// MockMessageItem 用于测试的模拟消息项
type MockMessageItem struct {
	mockTitle    string
	mockURL      string
	mockIsHot    bool
	mockIsNew    bool
	mockContent  string
	mockSource   string
	mockCategory string
}

func (m *MockMessageItem) Title() string {
	return m.mockTitle
}

func (m *MockMessageItem) URL() string {
	return m.mockURL
}

func (m *MockMessageItem) IsHot() bool {
	return m.mockIsHot
}

func (m *MockMessageItem) IsNew() bool {
	return m.mockIsNew
}

func (m *MockMessageItem) Content() string {
	return m.mockContent
}

func (m *MockMessageItem) GetSource() string {
	return m.mockSource
}

func (m *MockMessageItem) GetCategory() string {
	return m.mockCategory
}

// 测试创建邮件通知器
func TestNewEmailNotifier(t *testing.T) {
	// 创建禁用的配置
	disabledConfig := &EmailNotifierConfig{
		Enabled:  false,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	}

	// 测试创建禁用的通知器
	disabledNotifier, err := NewNotifier(disabledConfig)
	if err != nil {
		t.Fatalf("创建禁用的邮件通知器失败: %v", err)
	}
	if disabledNotifier.IsEnabled() {
		t.Error("禁用的通知器应该返回false")
	}

	// 创建启用的配置
	enabledConfig := &EmailNotifierConfig{
		Enabled:  true,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	}

	// 测试创建启用的通知器
	enabledNotifier, err := NewNotifier(enabledConfig)
	if err != nil {
		t.Fatalf("创建启用的邮件通知器失败: %v", err)
	}
	if !enabledNotifier.IsEnabled() {
		t.Error("启用的通知器应该返回true")
	}

	// 验证通知器名称
	if enabledNotifier.Name() != "email" {
		t.Errorf("通知器名称不匹配，期望'email'，实际得到'%s'", enabledNotifier.Name())
	}
}

// 测试IsEnabled方法
func TestEmailNotifierConfig_IsEnabled(t *testing.T) {
	// 测试启用状态
	enabledConfig := &EmailNotifierConfig{
		Enabled:  true,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	}

	if !enabledConfig.IsEnabled() {
		t.Error("启用的配置应该返回true")
	}

	// 测试禁用状态
	disabledConfig := &EmailNotifierConfig{
		Enabled:  false,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       []string{"recipient@example.com"},
	}

	if disabledConfig.IsEnabled() {
		t.Error("禁用的配置应该返回false")
	}
}
