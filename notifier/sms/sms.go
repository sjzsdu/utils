package sms

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sjzsdu/utils/notifier"
)

// SMSNotifierConfig 短信通知器配置
type SMSNotifierConfig struct {
	Enabled      bool     `yaml:"enabled" json:"enabled"`
	Provider     string   `yaml:"provider" json:"provider"`           // "aliyun", "tencent", "aws", "custom"
	PhoneNumbers []string `yaml:"phone_numbers" json:"phone_numbers"` // 接收短信的手机号码
	AccessKey    string   `yaml:"access_key" json:"access_key"`
	SecretKey    string   `yaml:"secret_key" json:"secret_key"`
	Region       string   `yaml:"region" json:"region"`
	TemplateID   string   `yaml:"template_id" json:"template_id"`
	Signature    string   `yaml:"signature" json:"signature"`
	CustomAPIURL string   `yaml:"custom_api_url" json:"custom_api_url"` // 自定义API地址
}

// IsEnabled 检查是否启用
func (c *SMSNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.Provider != "" && len(c.PhoneNumbers) > 0 &&
		((c.Provider != "custom" && c.AccessKey != "" && c.SecretKey != "") ||
			(c.Provider == "custom" && c.CustomAPIURL != ""))
}

// SMSNotifier 短信通知器
type SMSNotifier struct {
	config *SMSNotifierConfig
}

// NewNotifier 创建短信通知器
func NewNotifier(cfg *SMSNotifierConfig) (*SMSNotifier, error) {
	if cfg == nil {
		return nil, errors.New("短信配置为空")
	}

	// 验证手机号码格式
	for _, phone := range cfg.PhoneNumbers {
		if !isValidPhoneNumber(phone) {
			return nil, fmt.Errorf("手机号格式错误: %s", phone)
		}
	}

	return &SMSNotifier{
		config: cfg,
	}, nil
}

// Name 返回通知器名称
func (n *SMSNotifier) Name() string {
	return "sms"
}

// IsEnabled 检查是否启用
func (n *SMSNotifier) IsEnabled() bool {
	return n.config.IsEnabled()
}

// Send 发送通知
func (n *SMSNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
	result := &notifier.NotificationResult{
		Channel:    n.Name(),
		Status:     notifier.StatusPending,
		TotalCount: len(items),
		StartAt:    time.Now(),
	}

	if len(items) == 0 {
		result.Status = notifier.StatusSuccess
		result.EndAt = time.Now()
		return result, nil
	}

	// 短信内容需要简洁，只包含最重要的信息
	summary := notifier.FormatNotificationSummary(items)

	// 截取适当长度的内容（短信有长度限制）
	maxLength := 400
	if len(summary) > maxLength {
		summary = summary[:maxLength-3] + "..."
	}

	// 向每个手机号发送短信
	for _, phone := range n.config.PhoneNumbers {
		err := n.sendSMS(ctx, phone, summary)
		if err != nil {
			result.Status = notifier.StatusFailed
			result.Error = fmt.Errorf("向 %s 发送短信失败: %w", phone, err).Error()
			result.EndAt = time.Now()
			return result, err
		}
	}

	result.Status = notifier.StatusSuccess
	result.SuccessCount = len(items)
	result.EndAt = time.Now()
	return result, nil
}

// sendSMS 发送短信
func (n *SMSNotifier) sendSMS(ctx context.Context, phoneNumber, message string) error {
	switch n.config.Provider {
	case "aliyun":
		return n.sendAliyunSMS(ctx, phoneNumber, message)
	case "tencent":
		return n.sendTencentSMS(ctx, phoneNumber, message)
	case "aws":
		return n.sendAWSSMS(ctx, phoneNumber, message)
	case "custom":
		return n.sendCustomSMS(ctx, phoneNumber, message)
	default:
		return fmt.Errorf("不支持的短信服务提供商: %s", n.config.Provider)
	}
}

// sendAliyunSMS 发送阿里云短信
func (n *SMSNotifier) sendAliyunSMS(ctx context.Context, phoneNumber, message string) error {
	// 这里是阿里云短信服务的实现框架
	// 实际使用时需要集成阿里云SDK
	// 示例代码框架：
	// client := dysmsapi.NewClientWithAccessKey(n.config.Region, n.config.AccessKey, n.config.SecretKey)
	// request := dysmsapi.CreateSendSmsRequest()
	// request.PhoneNumbers = phoneNumber
	// request.SignName = n.config.Signature
	// request.TemplateCode = n.config.TemplateID
	// request.TemplateParam = fmt.Sprintf(`{"content":"%s"}`, message)
	// _, err := client.SendSms(request)

	// 模拟实现
	fmt.Printf("[阿里云] 向 %s 发送短信: %s\n", phoneNumber, message)
	return nil
}

// sendTencentSMS 发送腾讯云短信
func (n *SMSNotifier) sendTencentSMS(ctx context.Context, phoneNumber, message string) error {
	// 这里是腾讯云短信服务的实现框架
	// 实际使用时需要集成腾讯云SDK
	// 示例代码框架：
	// client := sms.NewClient(nil)
	// request := sms.NewSendSmsRequest()
	// request.SmsSdkAppId = n.config.AppID
	// request.SignName = n.config.Signature
	// request.TemplateId = n.config.TemplateID
	// request.PhoneNumberSet = []*string{&phoneNumber}
	// request.TemplateParamSet = []*string{&message}
	// _, err := client.SendSms(request)

	// 模拟实现
	fmt.Printf("[腾讯云] 向 %s 发送短信: %s\n", phoneNumber, message)
	return nil
}

// sendAWSSMS 发送AWS短信
func (n *SMSNotifier) sendAWSSMS(ctx context.Context, phoneNumber, message string) error {
	// 这里是AWS SNS短信服务的实现框架
	// 实际使用时需要集成AWS SDK
	// 示例代码框架：
	// sess := session.Must(session.NewSession())
	// svc := sns.New(sess, aws.NewConfig().WithRegion(n.config.Region))
	// params := &sns.PublishInput{
	//     Message:     aws.String(message),
	//     PhoneNumber: aws.String(phoneNumber),
	// }
	// _, err := svc.Publish(params)

	// 模拟实现
	fmt.Printf("[AWS] 向 %s 发送短信: %s\n", phoneNumber, message)
	return nil
}

// sendCustomSMS 发送自定义API短信
func (n *SMSNotifier) sendCustomSMS(ctx context.Context, phoneNumber, message string) error {
	// 这里是自定义API短信服务的实现
	// 实际使用时需要根据API文档进行HTTP请求

	// 模拟实现
	fmt.Printf("[自定义API] 向 %s 发送短信: %s\n", phoneNumber, message)
	return nil
}

// isValidPhoneNumber 验证手机号码格式
func isValidPhoneNumber(phone string) bool {
	// 简单的手机号验证，实际项目中可能需要更严格的验证
	phone = strings.TrimSpace(phone)
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}

	// 检查是否只包含数字和可能的国家代码前缀
	for _, ch := range phone {
		if !((ch >= '0' && ch <= '9') || ch == '+') {
			return false
		}
	}

	return true
}

// RegisterNotifier 注册短信通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("sms", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		smsConfig, ok := config.(*SMSNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNotifier(smsConfig)
	})
}
