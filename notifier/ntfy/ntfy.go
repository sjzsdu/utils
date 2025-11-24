package ntfy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sjzsdu/utils/notifier"
)

// 验证手机号格式
func validatePhoneNumber(phone string) bool {
	if phone == "" {
		return true
	}
	// 简单的手机号格式验证（国内手机号）
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return match
}

// NtfyNotifierConfig ntfy通知器配置
type NtfyNotifierConfig struct {
	Enabled   bool   `json:"enabled"`
	ServerURL string `json:"server_url,omitempty"`
	Topic     string `json:"topic"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	ClickURL  string `json:"click_url,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Proxy     string `json:"proxy,omitempty"`
}

// IsEnabled 检查是否启用
func (c *NtfyNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.Topic != ""
}

// NtfyNotifier ntfy通知器
type NtfyNotifier struct {
	config *NtfyNotifierConfig
	client *http.Client
}

// NtfyMessage ntfy消息结构
type NtfyMessage struct {
	Topic    string `json:"topic,omitempty"`
	Title    string `json:"title,omitempty"`
	Message  string `json:"message,omitempty"`
	Priority string `json:"priority,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Click    string `json:"click,omitempty"`
	Actions  string `json:"actions,omitempty"`
}

// NtfyResponse ntfy响应结构
type NtfyResponse struct {
	Id       string    `json:"id,omitempty"`
	Time     time.Time `json:"time,omitempty"`
	Topic    string    `json:"topic,omitempty"`
	Title    string    `json:"title,omitempty"`
	Message  string    `json:"message,omitempty"`
	Priority int       `json:"priority,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Click    string    `json:"click,omitempty"`
	Actions  []string  `json:"actions,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// NewNtfyNotifier 创建ntfy通知器
func NewNtfyNotifier(cfg *NtfyNotifierConfig) (*NtfyNotifier, error) {
	if cfg == nil {
		return nil, errors.New("ntfy配置为空")
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

	return &NtfyNotifier{
		config: cfg,
		client: client,
	}, nil
}

// Name 返回通知器名称
func (n *NtfyNotifier) Name() string {
	return "ntfy"
}

// IsEnabled 检查是否启用
func (n *NtfyNotifier) IsEnabled() bool {
	return n.config.Enabled && n.config.Topic != ""
}

// Send 发送通知
func (n *NtfyNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
	result := &notifier.NotificationResult{
		Channel:      n.Name(),
		Status:       notifier.StatusPending,
		TotalCount:   len(items),
		SuccessCount: 0,
		StartAt:      time.Now(),
	}

	if len(items) == 0 {
		result.Status = notifier.StatusSuccess
		result.EndAt = time.Now()
		return result, nil
	}

	// 格式化消息
	_, err := n.FormatMessage(items)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	result.Status = notifier.StatusSuccess // 修改为直接使用Success状态，因为我们没有StatusPartial

	// 为简化实现，这里不使用sendBatch函数，直接发送
	// 为所有消息格式化一条消息
	batchMessage, err := n.FormatMessage(items)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 获取API URL
	apiURL, err := n.getAPIURL()
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 创建请求
	req, err := n.createRequest(apiURL, batchMessage)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 添加认证信息
	if n.config.Username != "" && n.config.Password != "" {
		req.SetBasicAuth(n.config.Username, n.config.Password)
	}

	// 发送请求
	resp, err := n.client.Do(req.WithContext(ctx))
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = fmt.Errorf("发送请求失败: %w", err).Error()
		result.EndAt = time.Now()
		return result, err
	}
	defer resp.Body.Close()

	// 检查响应
	if err := n.checkResponse(resp); err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	result.SuccessCount = len(items)
	result.EndAt = time.Now()

	return result, nil
}

// FormatMessage 格式化消息
func (n *NtfyNotifier) FormatMessage(items []notifier.MessageItem) (string, error) {
	if len(items) == 0 {
		return "", errors.New("没有要发送的内容")
	}

	// 对于ntfy，我们需要将消息序列化为JSON
	message := &NtfyMessage{
		Topic:    n.config.Topic,
		Title:    notifier.FormatNotificationTitle(items),
		Priority: n.getPriority(),
		Tags:     n.getTags(items),
	}

	// 格式化消息内容
	message.Message = n.formatMessageContent(items)

	// 添加点击链接（如果有）
	if n.config.ClickURL != "" {
		message.Click = n.config.ClickURL
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("序列化消息失败: %w", err)
	}

	return string(jsonData), nil
}

// formatMessageContent 格式化消息内容
func (n *NtfyNotifier) formatMessageContent(items []notifier.MessageItem) string {
	var content strings.Builder

	// 添加摘要
	content.WriteString(fmt.Sprintf("%s\n\n", notifier.FormatNotificationSummary(items)))

	// 添加资讯列表（限制数量避免过长）
	maxItems := 5
	if len(items) < maxItems {
		maxItems = len(items)
	}

	for i := 0; i < maxItems; i++ {
		item := items[i]
		icon := "📄"

		// 资讯标题
		// 直接使用字符串切片截断，避免依赖外部函数
		maxTitleLength := 80
		title := item.Title()
		if len(title) > maxTitleLength {
			title = title[:maxTitleLength-3] + "..."
		}
		content.WriteString(fmt.Sprintf("%s %s\n", icon, title))

		// 资讯链接
		content.WriteString(fmt.Sprintf("链接: %s\n", item.URL()))

		// 资讯内容
		maxContentLength := 100
		contentStr := item.Content()
		if len(contentStr) > maxContentLength {
			contentStr = contentStr[:maxContentLength-3] + "..."
		}
		content.WriteString(fmt.Sprintf("内容: %s\n\n", contentStr))
	}

	// 如果有更多资讯，添加提示
	if len(items) > maxItems {
		content.WriteString(fmt.Sprintf("... 还有 %d 条资讯未显示", len(items)-maxItems))
	}

	return content.String()
}

// getAPIURL 获取API URL
func (n *NtfyNotifier) getAPIURL() (string, error) {
	baseURL := n.config.ServerURL
	if baseURL == "" {
		baseURL = "https://ntfy.sh"
	}

	// 确保baseURL以/结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// 构建完整的API URL
	apiURL := fmt.Sprintf("%s%s", baseURL, n.config.Topic)

	// 验证URL格式
	if _, err := url.Parse(apiURL); err != nil {
		return "", fmt.Errorf("无效的API URL: %w", err)
	}

	return apiURL, nil
}

// createRequest 创建HTTP请求
func (n *NtfyNotifier) createRequest(apiURL string, jsonData string) (*http.Request, error) {
	// 解析JSON数据
	var message NtfyMessage
	if err := json.Unmarshal([]byte(jsonData), &message); err != nil {
		return nil, fmt.Errorf("解析消息数据失败: %w", err)
	}

	// 构建表单数据
	data := url.Values{}
	data.Set("topic", message.Topic)
	data.Set("title", message.Title)
	data.Set("message", message.Message)
	if message.Priority != "" {
		data.Set("priority", message.Priority)
	}
	if message.Tags != "" {
		data.Set("tags", message.Tags)
	}
	if message.Click != "" {
		data.Set("click", message.Click)
	}

	// 创建请求
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

// checkResponse 检查响应
func (n *NtfyNotifier) checkResponse(resp *http.Response) error {
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response NtfyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// ntfy可能返回纯文本错误
		if strings.TrimSpace(string(body)) != "" {
			return fmt.Errorf("发送失败: %s", string(body))
		}
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if response.Error != "" {
		return fmt.Errorf("发送失败: %s", response.Error)
	}

	return nil
}

// getPriority 获取优先级
func (n *NtfyNotifier) getPriority() string {
	switch n.config.Priority {
	case "high":
		return "4"
	case "urgent":
		return "5"
	case "low":
		return "2"
	case "min":
		return "1"
	default:
		return "3" // 默认普通优先级
	}
}

// getTags 获取标签
func (n *NtfyNotifier) getTags(items []notifier.MessageItem) string {
	tags := []string{"information_source"}

	// 只返回基本标签，不再根据消息属性添加额外标签
	return strings.Join(tags, ",")
}

// GetMaxBatchSize 获取最大批次大小
func (n *NtfyNotifier) GetMaxBatchSize() int {
	// ntfy消息有大小限制，合理的批次大小
	return 5
}

// RegisterNotifier 注册ntfy通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("ntfy", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		ntfyConfig, ok := config.(*NtfyNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNtfyNotifier(ntfyConfig)
	})
}
