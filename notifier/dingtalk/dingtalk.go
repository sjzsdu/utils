package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

// DingtalkNotifierConfig 钉钉通知器配置
type DingtalkNotifierConfig struct {
	Enabled     bool   `json:"enabled"`
	WebhookURL  string `json:"webhook_url"`
	Secret      string `json:"secret,omitempty"`
	Proxy       string `json:"proxy,omitempty"`
	MessageType string `json:"message_type"`
}

// IsEnabled 检查是否启用
func (c *DingtalkNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.WebhookURL != ""
}

// DingtalkNotifier 钉钉通知器
type DingtalkNotifier struct {
	config *DingtalkNotifierConfig
	client *http.Client
}

// DingtalkMessage 钉钉消息结构
type DingtalkMessage struct {
	Msgtype string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text,omitempty"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown,omitempty"`
}

// NewNotifier 创建钉钉通知器
func NewNotifier(cfg *DingtalkNotifierConfig) (*DingtalkNotifier, error) {
	if cfg == nil {
		return nil, errors.New("钉钉配置为空")
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

	return &DingtalkNotifier{
		config: cfg,
		client: client,
	}, nil
}

// Name 返回通知器名称
func (n *DingtalkNotifier) Name() string {
	return "dingtalk"
}

// IsEnabled 检查是否启用
func (n *DingtalkNotifier) IsEnabled() bool {
	return n.config.IsEnabled()
}

// Send 发送通知
func (n *DingtalkNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
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

	// 构建请求URL
	requestURL, err := n.buildRequestURL()
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBufferString(messageBody))
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
		result.Error = fmt.Errorf("钉钉返回错误: %s (errcode: %.0f)", msg, errcode).Error()
		result.EndAt = time.Now()
		return result, err
	}

	result.Status = notifier.StatusSuccess
	result.SuccessCount = len(items)
	result.EndAt = time.Now()
	return result, nil
}

// formatTextMessage 格式化文本消息
func (n *DingtalkNotifier) formatTextMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	for i, item := range items {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title()))
		content.WriteString(fmt.Sprintf("   链接: %s\n", item.URL()))
		content.WriteString(fmt.Sprintf("   内容: %s\n", item.Content()))
		content.WriteString("\n")
	}

	msg := DingtalkMessage{
		Msgtype: "text",
	}
	msg.Text.Content = content.String()

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// formatMarkdownMessage 格式化Markdown消息
func (n *DingtalkNotifier) formatMarkdownMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", title))

	for i, item := range items {
		content.WriteString(fmt.Sprintf("## %d. [%s](%s)\n", i+1, item.Title(), item.URL()))
		content.WriteString(fmt.Sprintf("- **内容**: %s\n", item.Content()))
		content.WriteString("\n")
	}

	msg := DingtalkMessage{
		Msgtype: "markdown",
	}
	msg.Markdown.Title = title
	msg.Markdown.Text = content.String()

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// buildRequestURL 构建请求URL，添加签名
func (n *DingtalkNotifier) buildRequestURL() (string, error) {
	if n.config.Secret == "" {
		return n.config.WebhookURL, nil
	}

	// 生成签名
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, n.config.Secret)
	h := hmac.New(sha256.New, []byte(n.config.Secret))
	h.Write([]byte(stringToSign))
	signData := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 构建URL
	parsedURL, err := url.Parse(n.config.WebhookURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %w", err)
	}

	query := parsedURL.Query()
	query.Add("timestamp", fmt.Sprintf("%d", timestamp))
	query.Add("sign", url.QueryEscape(signData))
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// RegisterNotifier 注册钉钉通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("dingtalk", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		dingtalkConfig, ok := config.(*DingtalkNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNotifier(dingtalkConfig)
	})
}
