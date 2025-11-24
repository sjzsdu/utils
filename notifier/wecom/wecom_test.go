package wecom

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

// 测试创建企业微信通知器
func TestNewNotifier(t *testing.T) {
	// 创建禁用的配置
	disabledConfig := &WecomNotifierConfig{
		Enabled:    false,
		WebhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
	}

	// 测试创建禁用的通知器
	disabledNotifier, err := NewNotifier(disabledConfig)
	if err != nil {
		t.Fatalf("创建禁用的企业微信通知器失败: %v", err)
	}
	if disabledNotifier.IsEnabled() {
		t.Error("禁用的通知器应该返回false")
	}

	// 创建启用的配置
	enabledConfig := &WecomNotifierConfig{
		Enabled:     true,
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
		MessageType: "markdown",
	}

	// 测试创建启用的通知器
	enabledNotifier, err := NewNotifier(enabledConfig)
	if err != nil {
		t.Fatalf("创建启用的企业微信通知器失败: %v", err)
	}
	if !enabledNotifier.IsEnabled() {
		t.Error("启用的通知器应该返回true")
	}

	// 验证通知器名称
	if enabledNotifier.Name() != "wecom" {
		t.Errorf("通知器名称不匹配，期望'wecom'，实际得到'%s'", enabledNotifier.Name())
	}
}
