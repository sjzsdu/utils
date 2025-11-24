package telegram

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

// TelegramNotifierConfig Telegram通知器配置
type TelegramNotifierConfig struct {
	Enabled   bool   `json:"enabled"`
	BotToken  string `json:"bot_token"`
	ChatID    string `json:"chat_id"`
	Proxy     string `json:"proxy,omitempty"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// IsEnabled 检查是否启用
func (c *TelegramNotifierConfig) IsEnabled() bool {
	return c.Enabled && c.BotToken != "" && c.ChatID != ""
}

// TelegramNotifier Telegram通知器
type TelegramNotifier struct {
	config *TelegramNotifierConfig
	client *http.Client
}

// TelegramMessage Telegram消息结构
type TelegramMessage struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

// TelegramResponse Telegram响应结构
type TelegramResponse struct {
	OK          bool        `json:"ok"`
	Result      interface{} `json:"result,omitempty"`
	ErrorCode   int         `json:"error_code,omitempty"`
	Description string      `json:"description,omitempty"`
}

// NewTelegramNotifier 创建Telegram通知器
func NewTelegramNotifier(cfg *TelegramNotifierConfig) (*TelegramNotifier, error) {
	if cfg == nil {
		return nil, errors.New("Telegram配置为空")
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

	return &TelegramNotifier{
		config: cfg,
		client: client,
	}, nil
}

// Name 返回通知器名称
func (n *TelegramNotifier) Name() string {
	return "telegram"
}

// IsEnabled 检查是否启用
func (n *TelegramNotifier) IsEnabled() bool {
	return n.config.Enabled && n.config.BotToken != "" && n.config.ChatID != ""
}

// Send 发送通知
func (n *TelegramNotifier) Send(ctx context.Context, items []notifier.MessageItem) (*notifier.NotificationResult, error) {
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

	// 直接在Send方法中格式化消息
	var batchMessage string
	// 根据解析模式选择格式化方法
	if n.getParseMode() == "MarkdownV2" {
		batchMessage = n.formatMarkdownV2Message(items)
	} else if n.getParseMode() == "HTML" {
		batchMessage = n.formatHTMLMessage(items)
	} else {
		// 默认使用Markdown
		batchMessage = n.formatMarkdownMessage(items)
	}

	// 构建API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.config.BotToken)

	// 创建消息
	telegramMsg := &TelegramMessage{
		ChatID:    n.config.ChatID,
		Text:      batchMessage,
		ParseMode: n.getParseMode(),
		// TelegramConfig中没有DisableWebPagePreview字段，使用默认值true
		DisableWebPagePreview: true,
	}

	// 发送请求
	if err := n.sendRequest(ctx, apiURL, telegramMsg); err != nil {
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

// formatMarkdownMessage 格式化Markdown消息
func (n *TelegramNotifier) formatMarkdownMessage(items []notifier.MessageItem) string {
	var content strings.Builder

	// 添加标题
	content.WriteString(fmt.Sprintf("*%s*\n\n", notifier.FormatNotificationTitle(items)))

	// 添加摘要
	content.WriteString(fmt.Sprintf("_%s_\n\n", notifier.FormatNotificationSummary(items)))

	// 添加资讯列表
	for i, item := range items {
		icon := "📄"

		// 资讯标题
		content.WriteString(fmt.Sprintf("*%s %s*\n", icon, n.truncateText(item.Title(), 100)))

		// 资讯链接
		content.WriteString(fmt.Sprintf("[%s](%s)\n", "查看原文", item.URL()))

		// 资讯内容
		content.WriteString(fmt.Sprintf("%s\n", n.truncateText(item.Content(), 200)))

		// 非最后一条添加分隔线
		if i < len(items)-1 {
			content.WriteString("\n---\n\n")
		}
	}

	// 添加底部信息
	content.WriteString(fmt.Sprintf("\n*发送时间: %s*", time.Now().Format("2006-01-02 15:04:05")))

	return content.String()
}

// formatMarkdownV2Message 格式化MarkdownV2消息（需要转义特殊字符）
func (n *TelegramNotifier) formatMarkdownV2Message(items []notifier.MessageItem) string {
	var content strings.Builder

	// 添加标题
	content.WriteString(fmt.Sprintf("*%s*\n\n", n.escapeMarkdownV2(notifier.FormatNotificationTitle(items))))

	// 添加摘要
	content.WriteString(fmt.Sprintf("_%s_\n\n", n.escapeMarkdownV2(notifier.FormatNotificationSummary(items))))

	// 添加资讯列表
	for i, item := range items {
		icon := "📄"

		// 资讯标题
		content.WriteString(fmt.Sprintf("*%s %s*\n", icon, n.escapeMarkdownV2(n.truncateText(item.Title(), 100))))

		// 资讯链接
		content.WriteString(fmt.Sprintf("[%s](%s)\n", "查看原文", item.URL()))

		// 资讯内容
		content.WriteString(fmt.Sprintf("%s\n", n.escapeMarkdownV2(n.truncateText(item.Content(), 200))))

		// 非最后一条添加分隔线
		if i < len(items)-1 {
			content.WriteString("\n---\n\n")
		}
	}

	// 添加底部信息
	content.WriteString(fmt.Sprintf("\n*发送时间: %s*", n.escapeMarkdownV2(time.Now().Format("2006-01-02 15:04:05"))))

	return content.String()
}

// truncateText 截断文本
func (n *TelegramNotifier) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

// formatHTMLMessage 格式化HTML消息
func (n *TelegramNotifier) formatHTMLMessage(items []notifier.MessageItem) string {
	var content strings.Builder

	// 添加标题
	content.WriteString(fmt.Sprintf("<b>%s</b>\n\n", n.escapeHTML(notifier.FormatNotificationTitle(items))))

	// 添加摘要
	content.WriteString(fmt.Sprintf("<i>%s</i>\n\n", n.escapeHTML(notifier.FormatNotificationSummary(items))))

	// 添加资讯列表
	for i, item := range items {
		icon := "📄"

		// 资讯标题
		content.WriteString(fmt.Sprintf("<b>%s %s</b>\n", icon, n.escapeHTML(n.truncateText(item.Title(), 100))))

		// 资讯链接
		content.WriteString(fmt.Sprintf("<a href='%s'>查看原文</a>\n", item.URL()))

		// 资讯内容
		content.WriteString(fmt.Sprintf("<p>%s</p>\n", n.escapeHTML(n.truncateText(item.Content(), 200))))

		// 非最后一条添加分隔线
		if i < len(items)-1 {
			content.WriteString("\n<hr>\n\n")
		}
	}

	// 添加底部信息
	content.WriteString(fmt.Sprintf("\n<b>发送时间: %s</b>", n.escapeHTML(time.Now().Format("2006-01-02 15:04:05"))))

	return content.String()
}

// sendRequest 发送请求
func (n *TelegramNotifier) sendRequest(ctx context.Context, url string, message *TelegramMessage) error {
	// 构建表单数据
	data := make(map[string]string)
	data["chat_id"] = message.ChatID
	data["text"] = message.Text
	if message.ParseMode != "" {
		data["parse_mode"] = message.ParseMode
	}
	if message.DisableWebPagePreview {
		data["disable_web_page_preview"] = "true"
	}

	// 序列化为JSON而不是表单
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(dataJSON))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := n.client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

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
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if !response.OK {
		return fmt.Errorf("发送失败: %s (错误码: %d)", response.Description, response.ErrorCode)
	}

	return nil
}

// getParseMode 获取解析模式
func (n *TelegramNotifier) getParseMode() string {
	// TelegramConfig中没有ParseMode字段，使用默认值HTML
	return "HTML"
}

// escapeMarkdownV2 转义MarkdownV2特殊字符
func (n *TelegramNotifier) escapeMarkdownV2(text string) string {
	// Telegram MarkdownV2需要转义的特殊字符
	specialChars := []rune{'_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'}
	escaped := strings.Builder{}

	for _, char := range text {
		for _, special := range specialChars {
			if char == special {
				escaped.WriteRune('\\')
				break
			}
		}
		escaped.WriteRune(char)
	}

	return escaped.String()
}

// escapeHTML 转义HTML特殊字符
func (n *TelegramNotifier) escapeHTML(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(text)
}

// GetMaxBatchSize 获取最大批次大小
func (n *TelegramNotifier) GetMaxBatchSize() int {
	// Telegram消息有字符限制，合理的批次大小
	return 10
}

// RegisterNotifier 注册Telegram通知器
func RegisterNotifier(registry *notifier.NotifierRegistry) {
	registry.Register("telegram", func(config notifier.NotifierConfig) (notifier.Notifier, error) {
		telegramConfig, ok := config.(*TelegramNotifierConfig)
		if !ok {
			return nil, errors.New("配置类型错误")
		}
		return NewTelegramNotifier(telegramConfig)
	})
}
