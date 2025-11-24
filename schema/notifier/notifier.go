package notifier

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sjzsdu/utils/notifier"
	"github.com/sjzsdu/utils/notifier/dingtalk"
	"github.com/sjzsdu/utils/notifier/email"
	"github.com/sjzsdu/utils/notifier/feishu"
	"github.com/sjzsdu/utils/notifier/ntfy"
	"github.com/sjzsdu/utils/notifier/sms"
	"github.com/sjzsdu/utils/notifier/webhook"
	"github.com/sjzsdu/utils/notifier/wecom"
)

// ManagerSchema 管理通知器的schema
type ManagerSchema struct {
	config Config
}

// NewManagerSchema 创建ManagerSchema实例
func NewManagerSchema() *ManagerSchema {
	return &ManagerSchema{}
}

// LoadFromFile 从配置文件加载配置
func (s *ManagerSchema) LoadFromFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(content, &s.config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// LoadFromBytes 从字节数组加载配置
func (s *ManagerSchema) LoadFromBytes(data []byte) error {
	if err := yaml.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}

// CreateNotifierManager 根据配置创建NotifierManager
func (s *ManagerSchema) CreateNotifierManager() (*notifier.NotifierManager, error) {
	manager, err := notifier.NewNotifierManager()
	if err != nil {
		return nil, err
	}

	// 创建并注册钉钉通知器
	if s.config.Dingtalk != nil {
		dingtalkNotifier, err := dingtalk.NewNotifier(s.config.Dingtalk)
		if err != nil {
			return nil, fmt.Errorf("创建钉钉通知器失败: %w", err)
		}
		manager.RegisterNotifier("dingtalk", dingtalkNotifier)
	}

	// 创建并注册邮件通知器
	if s.config.Email != nil {
		emailNotifier, err := email.NewNotifier(s.config.Email)
		if err != nil {
			return nil, fmt.Errorf("创建邮件通知器失败: %w", err)
		}
		manager.RegisterNotifier("email", emailNotifier)
	}

	// 创建并注册飞书通知器
	if s.config.Feishu != nil {
		feishuNotifier, err := feishu.NewNotifier(s.config.Feishu)
		if err != nil {
			return nil, fmt.Errorf("创建飞书通知器失败: %w", err)
		}
		manager.RegisterNotifier("feishu", feishuNotifier)
	}

	// 创建并注册NTFY通知器
	if s.config.NTFY != nil {
		ntfyNotifier, err := ntfy.NewNtfyNotifier(s.config.NTFY)
		if err != nil {
			return nil, fmt.Errorf("创建NTFY通知器失败: %w", err)
		}
		manager.RegisterNotifier("ntfy", ntfyNotifier)
	}

	// 创建并注册短信通知器
	if s.config.SMS != nil {
		smsNotifier, err := sms.NewNotifier(s.config.SMS)
		if err != nil {
			return nil, fmt.Errorf("创建短信通知器失败: %w", err)
		}
		manager.RegisterNotifier("sms", smsNotifier)
	}

	// 创建并注册Webhook通知器
	if s.config.Webhook != nil {
		webhookNotifier, err := webhook.NewNotifier(s.config.Webhook)
		if err != nil {
			return nil, fmt.Errorf("创建Webhook通知器失败: %w", err)
		}
		manager.RegisterNotifier("webhook", webhookNotifier)
	}

	// 创建并注册企业微信通知器
	if s.config.Wecom != nil {
		wecomNotifier, err := wecom.NewNotifier(s.config.Wecom)
		if err != nil {
			return nil, fmt.Errorf("创建企业微信通知器失败: %w", err)
		}
		manager.RegisterNotifier("wecom", wecomNotifier)
	}

	return manager, nil
}

// LoadAndCreateNotifierManager 从配置文件加载配置并创建NotifierManager
func LoadAndCreateNotifierManager(filePath string) (*notifier.NotifierManager, error) {
	schema := NewManagerSchema()
	if err := schema.LoadFromFile(filePath); err != nil {
		return nil, err
	}

	return schema.CreateNotifierManager()
}
