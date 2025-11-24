package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sjzsdu/utils/notifier"
)

// WecomNotifierConfig 企业微信通知器配置
type WecomNotifierConfig struct {
	Enabled     bool   `json:"enabled"`
	WebhookURL  string `json:"webhook_url"`
	AgentID     string `json:"agent_id,omitempty"`
	CorpID      string `json:"corp_id,omitempty"`
	CorpSecret  string `json:"corp_secret,omitempty"`
	ToUser      string `json:"to_user,omitempty"`
	ToParty     string `json:"to_party,omitempty"`
	ToTag       string `json:"to_tag,omitempty"`
	Proxy       string `json:"proxy,omitempty"`
	MessageType string `json:"message_type"`
}

// IsEnabled 检查是否启用
func (c *WecomNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.WebhookURL != ""
}

// WecomNotifier 企业微信通知器
type WecomNotifier struct {
	config *WecomNotifierConfig
	client *http.Client
}

// WecomMessage 企业微信消息结构
type WecomMessage struct {
	Msgtype  string                `json:"msgtype"`
	ToUser   string                `json:"touser,omitempty"`
	ToParty  string                `json:"toparty,omitempty"`
	ToTag    string                `json:"totag,omitempty"`
	Text     *WecomTextMessage     `json:"text,omitempty"`
	Markdown *WecomMarkdownMessage `json:"markdown,omitempty"`
}

// WecomTextMessage 企业微信文本消息
type WecomTextMessage struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// WecomMarkdownMessage 企业微信Markdown消息
type WecomMarkdownMessage struct {
	Content string `json:"content"`
}

// NewNotifier 创建企业微信通知器
func NewNotifier(cfg *WecomNotifierConfig) (*WecomNotifier, error) {
	if cfg == nil {
		return nil, errors.New("企业微信配置为空")
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 配置代理
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("代理配置无效: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return &WecomNotifier{
		config: cfg,
		client: client,
	}, nil
}

// Name 返回通知器名称
func (n *WecomNotifier) Name() string {
	return "wecom"
}

// IsEnabled 检查是否启用
func (n *WecomNotifier) IsEnabled() bool {
	return n.config.IsEnabled()
}

// Send 发送通知
func (n *WecomNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
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

	// 直接在Send方法中格式化消息
	var messageBody string
	var err error

	// 根据消息类型格式化内容
	switch n.config.MessageType {
	case "markdown":
		messageBody, err = n.formatMarkdownMessage(title, items)
	default:
		messageBody, err = n.formatTextMessage(title, items)
	}

	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", n.config.WebhookURL, bytes.NewBufferString(messageBody))
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("创建请求失败: %w", err).Error()
		result.EndAt = time.Now()
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := n.client.Do(req)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("发送请求失败: %w", err).Error()
		result.EndAt = time.Now()
		return result, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("读取响应失败: %w", err).Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody)).Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 解析响应
	var respData map[string]interface{}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("解析响应失败: %w", err).Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 检查错误码
	if errcode, ok := respData["errcode"].(float64); ok && errcode != 0 {
		msg := "未知错误"
		if message, ok := respData["errmsg"].(string); ok {
			msg = message
		}
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("企业微信返回错误: %s (errcode: %.0f)", msg, errcode).Error()
		result.EndAt = time.Now()
		return result, err
	}

	result.Status = notifier.StatusSuccess
	result.SuccessCount = len(items)
	result.EndAt = time.Now()
	return result, nil
}

// formatTextMessage 格式化文本消息
func (n *WecomNotifier) formatTextMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	for i, item := range items {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title()))
		content.WriteString(fmt.Sprintf("   链接: %s\n", item.URL()))
		content.WriteString(fmt.Sprintf("   内容: %s\n", item.Content()))
		content.WriteString("\n")
	}

	msg := WecomMessage{
		Msgtype: "text",
		ToUser:  n.config.ToUser,
		ToParty: n.config.ToParty,
		ToTag:   n.config.ToTag,
		Text: &WecomTextMessage{
			Content: content.String(),
		},
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// formatMarkdownMessage 格式化Markdown消息
func (n *WecomNotifier) formatMarkdownMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", title))

	for i, item := range items {
		content.WriteString(fmt.Sprintf("## %d. [%s](%s)\n", i+1, item.Title(), item.URL()))
		content.WriteString(fmt.Sprintf("- **内容**: %s\n", item.Content()))
		content.WriteString("\n")
	}

	msg := WecomMessage{
		Msgtype:  "markdown",
		ToUser:   n.config.ToUser,
		ToParty:  n.config.ToParty,
		ToTag:    n.config.ToTag,
		Markdown: &WecomMarkdownMessage{Content: content.String()},
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// RegisterNotifier 注册企业微信通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("wecom", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		wecomConfig, ok := config.(*WecomNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNotifier(wecomConfig)
	})
}
