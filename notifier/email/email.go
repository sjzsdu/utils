package email

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/sjzsdu/utils/notifier"
)

// EmailNotifierConfig 邮件通知器配置
type EmailNotifierConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	SMTPHost    string   `yaml:"smtp_host" json:"smtp_host"`
	SMTPPort    int      `yaml:"smtp_port" json:"smtp_port"`
	Username    string   `yaml:"username" json:"username"`
	Password    string   `yaml:"password" json:"password"`
	From        string   `yaml:"from" json:"from"`
	To          []string `yaml:"to" json:"to"`
	CC          []string `yaml:"cc,omitempty" json:"cc,omitempty"`
	BCC         []string `yaml:"bcc,omitempty" json:"bcc,omitempty"`
	UseTLS      bool     `yaml:"use_tls" json:"use_tls"`
	UseSSL      bool     `yaml:"use_ssl" json:"use_ssl"`
	MessageType string   `yaml:"message_type" json:"message_type"`
}

// IsEnabled 检查是否启用
func (c *EmailNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.SMTPHost != "" && c.Username != "" && c.Password != "" && len(c.To) > 0
}

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	config *EmailNotifierConfig
}

// NewNotifier 创建邮件通知器
func NewNotifier(cfg *EmailNotifierConfig) (*EmailNotifier, error) {
	if cfg == nil {
		return nil, errors.New("邮件配置为空")
	}

	// 验证邮箱格式
	if _, err := mail.ParseAddress(cfg.From); err != nil {
		return nil, fmt.Errorf("发件人邮箱格式错误: %w", err)
	}

	for _, to := range cfg.To {
		if _, err := mail.ParseAddress(to); err != nil {
			return nil, fmt.Errorf("收件人邮箱格式错误: %w", err)
		}
	}

	return &EmailNotifier{
		config: cfg,
	}, nil
}

// Name 返回通知器名称
func (n *EmailNotifier) Name() string {
	return "email"
}

// IsEnabled 检查是否启用
func (n *EmailNotifier) IsEnabled() bool {
	return n.config.IsEnabled()
}

// Send 发送通知
func (n *EmailNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
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

	// 格式化标题
	title := notifier.FormatNotificationTitle(items)
	var messageBody string
	var err error

	// 根据消息类型格式化内容
	switch n.config.MessageType {
	case "html":
		messageBody, err = n.formatHTMLMessage(title, items)
	default:
		messageBody, err = n.formatTextMessage(title, items)
	}

	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 发送邮件
	if err := n.sendEmail(messageBody); err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	result.Status = notifier.StatusSuccess
	result.SuccessCount = len(items)
	result.EndAt = time.Now()
	return result, nil
}

// formatTextMessage 格式化文本消息
func (n *EmailNotifier) formatTextMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Subject: %s\n", title))
	content.WriteString(fmt.Sprintf("From: %s\n", n.config.From))
	content.WriteString(fmt.Sprintf("To: %s\n", strings.Join(n.config.To, ", ")))
	if len(n.config.CC) > 0 {
		content.WriteString(fmt.Sprintf("CC: %s\n", strings.Join(n.config.CC, ", ")))
	}
	content.WriteString("Content-Type: text/plain; charset=utf-8\n")
	content.WriteString("\n")

	content.WriteString(notifier.FormatNotificationSummary(items))
	content.WriteString("\n\n")

	for i, item := range items {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title()))
		content.WriteString(fmt.Sprintf("   链接: %s\n", item.URL()))
		content.WriteString(fmt.Sprintf("   内容: %s\n", item.Content()))
		content.WriteString("\n")
	}

	return content.String(), nil
}

// formatHTMLMessage 格式化HTML消息
func (n *EmailNotifier) formatHTMLMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Subject: %s\n", title))
	content.WriteString(fmt.Sprintf("From: %s\n", n.config.From))
	content.WriteString(fmt.Sprintf("To: %s\n", strings.Join(n.config.To, ", ")))
	if len(n.config.CC) > 0 {
		content.WriteString(fmt.Sprintf("CC: %s\n", strings.Join(n.config.CC, ", ")))
	}
	content.WriteString("Content-Type: text/html; charset=utf-8\n")
	content.WriteString("\n")

	content.WriteString("<html><body>")
	content.WriteString(fmt.Sprintf("<h1>%s</h1>", title))
	content.WriteString(fmt.Sprintf("<p>%s</p>", notifier.FormatNotificationSummary(items)))

	content.WriteString("<table border='1' cellpadding='5' cellspacing='0' style='border-collapse: collapse; width: 100%;'>")
	content.WriteString("<tr style='background-color: #f2f2f2;'>")
	content.WriteString("<th>序号</th><th>标题</th><th>内容</th>")
	content.WriteString("</tr>")

	for i, item := range items {
		content.WriteString("<tr>")
		content.WriteString(fmt.Sprintf("<td>%d</td>", i+1))
		content.WriteString(fmt.Sprintf("<td><a href='%s'>%s</a></td>", item.URL(), item.Title()))
		content.WriteString(fmt.Sprintf("<td>%s</td>", item.Content()))
		content.WriteString("</tr>")
	}

	content.WriteString("</table>")
	content.WriteString("</body></html>")

	return content.String(), nil
}

// sendEmail 发送邮件
func (n *EmailNotifier) sendEmail(emailContent string) error {
	// 准备收件人列表
	recipients := make([]string, 0)
	recipients = append(recipients, n.config.To...)
	recipients = append(recipients, n.config.CC...)
	recipients = append(recipients, n.config.BCC...)

	// 构建SMTP地址
	smtpAddr := fmt.Sprintf("%s:%d", n.config.SMTPHost, n.config.SMTPPort)

	// 认证信息
	auth := smtp.PlainAuth("", n.config.Username, n.config.Password, n.config.SMTPHost)

	// 发送邮件
	return smtp.SendMail(smtpAddr, auth, n.config.From, recipients, []byte(emailContent))
}

// RegisterNotifier 注册邮件通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("email", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		emailConfig, ok := config.(*EmailNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNotifier(emailConfig)
	})
}
