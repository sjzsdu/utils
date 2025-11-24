package feishu

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

// FeishuNotifierConfig 飞书通知器配置
type FeishuNotifierConfig struct {
	Enabled     bool   `json:"enabled"`
	WebhookURL  string `json:"webhook_url"`
	Secret      string `json:"secret,omitempty"`
	Proxy       string `json:"proxy,omitempty"`
	MessageType string `json:"message_type"`
}

// IsEnabled 检查是否启用
func (c *FeishuNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.WebhookURL != ""
}

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	config *FeishuNotifierConfig
	client *http.Client
}

// FeishuMessage 飞书消息结构
type FeishuMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text     string `json:"text,omitempty"`
		Post     string `json:"post,omitempty"`
		Markdown string `json:"markdown,omitempty"`
	} `json:"content"`
}

// FeishuPostContent 飞书富文本内容
type FeishuPostContent struct {
	ZhCN struct {
		Title   string                `json:"title"`
		Content [][]map[string]string `json:"content"`
	} `json:"zh_cn"`
}

// NewNotifier 创建飞书通知器
func NewNotifier(cfg *FeishuNotifierConfig) (*FeishuNotifier, error) {
	if cfg == nil {
		return nil, errors.New("飞书配置为空")
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

	return &FeishuNotifier{
		config: cfg,
		client: client,
	}, nil
}

// Name 返回通知器名称
func (n *FeishuNotifier) Name() string {
	return "feishu"
}

// IsEnabled 检查是否启用
func (n *FeishuNotifier) IsEnabled() bool {
	return n.config.IsEnabled()
}

// Send 发送通知
func (n *FeishuNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
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
	case "post":
		messageBody, err = n.formatPostMessage(title, items)
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
	if code, ok := respData["code"].(float64); ok && code != 0 {
		msg := "未知错误"
		if message, ok := respData["msg"].(string); ok {
			msg = message
		}
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("飞书返回错误: %s (code: %.0f)", msg, code).Error()
		result.EndAt = time.Now()
		return result, err
	}

	result.Status = notifier.StatusSuccess
	result.SuccessCount = len(items)
	result.EndAt = time.Now()
	return result, nil
}

// formatTextMessage 格式化文本消息
func (n *FeishuNotifier) formatTextMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	for i, item := range items {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title()))
		content.WriteString(fmt.Sprintf("   链接: %s\n", item.URL()))
		content.WriteString(fmt.Sprintf("   内容: %s\n", item.Content()))
		content.WriteString("\n")
	}

	msg := FeishuMessage{
		MsgType: "text",
	}
	msg.Content.Text = content.String()

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// formatMarkdownMessage 格式化Markdown消息
func (n *FeishuNotifier) formatMarkdownMessage(title string, items []notifier.MessageItem) (string, error) {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", title))

	for i, item := range items {
		content.WriteString(fmt.Sprintf("## %d. [%s](%s)\n", i+1, item.Title(), item.URL()))
		content.WriteString(fmt.Sprintf("- **内容**: %s\n", item.Content()))
		content.WriteString("\n")
	}

	msg := FeishuMessage{
		MsgType: "markdown",
	}
	msg.Content.Markdown = content.String()

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// formatPostMessage 格式化富文本消息
func (n *FeishuNotifier) formatPostMessage(title string, items []notifier.MessageItem) (string, error) {
	post := FeishuPostContent{}
	post.ZhCN.Title = title
	post.ZhCN.Content = make([][]map[string]string, 0)

	// 添加摘要行
	summaryRow := []map[string]string{
		{
			"tag":  "text",
			"text": notifier.FormatNotificationSummary(items),
		},
	}
	post.ZhCN.Content = append(post.ZhCN.Content, summaryRow)

	// 添加空行
	post.ZhCN.Content = append(post.ZhCN.Content, []map[string]string{{"tag": "text", "text": ""}})

	// 添加消息项
	for i, item := range items {
		row := []map[string]string{
			{
				"tag":  "text",
				"text": fmt.Sprintf("%d. ", i+1),
			},
			{
				"tag":  "a",
				"text": item.Title(),
				"href": item.URL(),
			},
		}
		post.ZhCN.Content = append(post.ZhCN.Content, row)

		// 添加内容行
		contentRow := []map[string]string{
			{
				"tag":  "text",
				"text": fmt.Sprintf("   内容: %s", item.Content()),
			},
		}
		post.ZhCN.Content = append(post.ZhCN.Content, contentRow)

		// 添加空行
		post.ZhCN.Content = append(post.ZhCN.Content, []map[string]string{{"tag": "text", "text": ""}})
	}

	postJSON, err := json.Marshal(post)
	if err != nil {
		return "", fmt.Errorf("序列化富文本内容失败: %w", err)
	}

	msg := FeishuMessage{
		MsgType: "post",
	}
	msg.Content.Post = string(postJSON)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// buildRequestURL 构建请求URL，添加签名
func (n *FeishuNotifier) buildRequestURL() (string, error) {
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
	query.Add("sign", signData)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// RegisterNotifier 注册飞书通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("feishu", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		feishuConfig, ok := config.(*FeishuNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNotifier(feishuConfig)
	})
}
