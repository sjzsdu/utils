package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sjzsdu/utils/notifier"
)

// WebhookNotifierConfig Webhook通知器配置
type WebhookNotifierConfig struct {
	Enabled       bool              `yaml:"enabled" json:"enabled"`
	URL           string            `yaml:"url" json:"url"`
	Method        string            `yaml:"method" json:"method"`                       // GET, POST, PUT 等
	Headers       map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"` // 自定义HTTP头
	Timeout       int               `yaml:"timeout,omitempty" json:"timeout,omitempty"` // 超时时间（秒）
	RetryCount    int               `yaml:"retry_count,omitempty" json:"retry_count,omitempty"`
	RetryInterval int               `yaml:"retry_interval,omitempty" json:"retry_interval,omitempty"` // 重试间隔（秒）
	ContentType   string            `yaml:"content_type,omitempty" json:"content_type,omitempty"`
	Secret        string            `yaml:"secret,omitempty" json:"secret,omitempty"` // 用于签名的密钥
}

// IsEnabled 检查是否启用
func (c *WebhookNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.URL != ""
}

// WebhookNotifier Webhook通知器
type WebhookNotifier struct {
	config *WebhookNotifierConfig
	client *http.Client
}

// NewNotifier 创建Webhook通知器
func NewNotifier(cfg *WebhookNotifierConfig) (*WebhookNotifier, error) {
	if cfg == nil {
		return nil, errors.New("Webhook配置为空")
	}

	if cfg.URL == "" {
		return nil, errors.New("Webhook URL不能为空")
	}

	// 设置默认值
	if cfg.Method == "" {
		cfg.Method = "POST"
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 // 默认10秒
	}

	if cfg.RetryCount < 0 {
		cfg.RetryCount = 2 // 默认重试2次
	}

	if cfg.RetryInterval <= 0 {
		cfg.RetryInterval = 2 // 默认2秒
	}

	if cfg.ContentType == "" {
		cfg.ContentType = "application/json"
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	return &WebhookNotifier{
		config: cfg,
		client: client,
	}, nil
}

// Name 返回通知器名称
func (n *WebhookNotifier) Name() string {
	return "webhook"
}

// IsEnabled 检查是否启用
func (n *WebhookNotifier) IsEnabled() bool {
	return n.config.IsEnabled()
}

// Send 发送通知
func (n *WebhookNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
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

	// 构建payload
	payload, err := n.buildPayload(items)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
		result.EndAt = time.Now()
		return result, err
	}

	// 发送请求
	err = n.sendRequest(ctx, payload)
	if err != nil {
		result.Status = notifier.StatusFailed
		result.Error = err.Error()
	} else {
		result.Status = notifier.StatusSuccess
		result.SuccessCount = len(items)
	}
	result.EndAt = time.Now()
	return result, err
}

// WebhookPayload Webhook发送的负载结构
type WebhookPayload struct {
	Title     string                 `json:"title"`
	Summary   string                 `json:"summary"`
	Items     []WebhookMessageItem   `json:"items"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookMessageItem Webhook消息项结构
type WebhookMessageItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// buildPayload 构建请求payload
func (n *WebhookNotifier) buildPayload(items []notifier.MessageItem) ([]byte, error) {
	// 创建消息项数组
	webhookItems := make([]WebhookMessageItem, 0, len(items))
	for _, item := range items {
		webhookItems = append(webhookItems, WebhookMessageItem{
			Title:   item.Title(),
			URL:     item.URL(),
			Content: item.Content(),
		})
	}

	// 创建payload
	payload := WebhookPayload{
		Title:     notifier.FormatNotificationTitle(items),
		Summary:   notifier.FormatNotificationSummary(items),
		Items:     webhookItems,
		Timestamp: time.Now().Unix(),
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %w", err)
	}

	return jsonData, nil
}

// sendRequest 发送HTTP请求
func (n *WebhookNotifier) sendRequest(ctx context.Context, payload []byte) error {
	var err error
	for attempt := 0; attempt <= n.config.RetryCount; attempt++ {
		if attempt > 0 {
			// 等待重试间隔
			time.Sleep(time.Duration(n.config.RetryInterval) * time.Second)
		}

		err = n.doRequest(ctx, payload)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("达到最大重试次数: %w", err)
}

// doRequest 执行单次HTTP请求
func (n *WebhookNotifier) doRequest(ctx context.Context, payload []byte) error {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, n.config.Method, n.config.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", n.config.ContentType)
	req.Header.Set("User-Agent", "notifier-webhook-client/1.0")

	// 添加自定义请求头
	for key, value := range n.config.Headers {
		req.Header.Set(key, value)
	}

	// 如果配置了密钥，添加签名头
	if n.config.Secret != "" {
		signature := n.generateSignature(payload)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// 发送请求
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateSignature 生成请求签名
func (n *WebhookNotifier) generateSignature(payload []byte) string {
	// 这里实现签名逻辑，例如HMAC SHA256
	// 实际项目中应根据接收端的要求实现正确的签名算法
	// 这里返回一个简单的示例
	return fmt.Sprintf("sha256=%x", payload)
}

// FormatMessage 格式化消息
func (n *WebhookNotifier) FormatMessage(items []notifier.MessageItem) (string, error) {
	// 构建payload
	payload, err := n.buildPayload(items)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// RegisterNotifier 注册Webhook通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("webhook", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		webhookConfig, ok := config.(*WebhookNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewNotifier(webhookConfig)
	})
}
