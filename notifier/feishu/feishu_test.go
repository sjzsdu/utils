package feishu

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

// 测试创建飞书通知器
func TestNewNotifier(t *testing.T) {
	// 创建禁用的配置
	disabledConfig := &FeishuNotifierConfig{
		Enabled:    false,
		WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
	}

	// 测试创建禁用的通知器
	disabledNotifier, err := NewNotifier(disabledConfig)
	if err != nil {
		t.Fatalf("创建禁用的飞书通知器失败: %v", err)
	}
	if disabledNotifier.IsEnabled() {
		t.Error("禁用的通知器应该返回false")
	}

	// 创建启用的配置
	enabledConfig := &FeishuNotifierConfig{
		Enabled:     true,
		WebhookURL:  "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
		MessageType: "markdown",
	}

	// 测试创建启用的通知器
	enabledNotifier, err := NewNotifier(enabledConfig)
	if err != nil {
		t.Fatalf("创建启用的飞书通知器失败: %v", err)
	}
	if !enabledNotifier.IsEnabled() {
		t.Error("启用的通知器应该返回true")
	}

	// 验证通知器名称
	if enabledNotifier.Name() != "feishu" {
		t.Errorf("通知器名称不匹配，期望'feishu'，实际得到'%s'", enabledNotifier.Name())
	}
}
