package webhook

import (
	"testing"

	"github.com/sjzsdu/utils/notifier"
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

// 测试创建Webhook通知器
func TestNewNotifier(t *testing.T) {
	// 创建禁用的配置
	disabledConfig := &WebhookNotifierConfig{
		Enabled: false,
		URL:     "https://example.com/webhook",
		Method:  "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 测试创建禁用的通知器
	disabledNotifier, err := NewNotifier(disabledConfig)
	if err != nil {
		t.Fatalf("创建禁用的Webhook通知器失败: %v", err)
	}
	if disabledNotifier.IsEnabled() {
		t.Error("禁用的通知器应该返回false")
	}

	// 创建启用的配置
	enabledConfig := &WebhookNotifierConfig{
		Enabled: true,
		URL:     "https://example.com/webhook",
		Method:  "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Timeout:    5,
		RetryCount: 3,
	}

	// 测试创建启用的通知器
	enabledNotifier, err := NewNotifier(enabledConfig)
	if err != nil {
		t.Fatalf("创建启用的Webhook通知器失败: %v", err)
	}
	if !enabledNotifier.IsEnabled() {
		t.Error("启用的通知器应该返回true")
	}

	// 验证通知器名称
	if enabledNotifier.Name() != "webhook" {
		t.Errorf("通知器名称不匹配，期望'webhook'，实际得到'%s'", enabledNotifier.Name())
	}

	// 测试无效配置
	invalidConfig := &WebhookNotifierConfig{
		Enabled: true,
		// 缺少URL
		Method: "POST",
	}

	_, err = NewNotifier(invalidConfig)
	if err == nil {
		t.Error("使用无效配置创建通知器应该返回错误")
	}
}

// 测试格式化消息 - JSON格式
func TestFormatJSONMessage(t *testing.T) {
	// 创建测试配置 - JSON格式
	config := &WebhookNotifierConfig{
		Enabled: true,
		URL:     "https://example.com/webhook",
		Method:  "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 创建通知器
	webhookNotifier, _ := NewNotifier(config)

	// 创建测试消息项
	items := []notifier.MessageItem{
		&MockMessageItem{
			mockTitle:   "测试标题1",
			mockURL:     "https://example.com/1",
			mockContent: "测试内容1",
		},
		&MockMessageItem{
			mockTitle:   "测试标题2",
			mockURL:     "https://example.com/2",
			mockContent: "测试内容2",
		},
	}

	// 格式化消息
	message, err := webhookNotifier.FormatMessage(items)
	if err != nil {
		t.Fatalf("格式化JSON消息失败: %v", err)
	}

	if message == "" {
		t.Error("格式化的JSON消息为空")
	}

	// 检查是否以JSON格式开头和结尾
	if message != "" && message[:1] != "{" && message[len(message)-1:] != "}" {
		t.Error("JSON格式的消息应该以'{'开头并以'}'结尾")
	}
}

// 测试格式化消息 - 文本格式
func TestFormatTextMessage(t *testing.T) {
	// 创建测试配置 - 文本格式
	config := &WebhookNotifierConfig{
		Enabled: true,
		URL:     "https://example.com/webhook",
		Method:  "POST",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}

	// 创建通知器
	webhookNotifier, _ := NewNotifier(config)

	// 创建测试消息项
	items := []notifier.MessageItem{
		&MockMessageItem{
			mockTitle:   "测试标题",
			mockURL:     "https://example.com",
			mockContent: "测试内容",
		},
	}

	// 格式化消息
	message, err := webhookNotifier.FormatMessage(items)
	if err != nil {
		t.Fatalf("格式化文本消息失败: %v", err)
	}

	if message == "" {
		t.Error("格式化的文本消息为空")
	}
}
