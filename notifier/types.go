package notifier

import (
	"context"
	"time"
)

// MessageItem 消息项接口
type MessageItem interface {
	// Title 获取标题
	Title() string
	// URL 获取链接
	URL() string
	// Content 获取内容
	Content() string
}

// NotifierConfig 通知器配置接口
type NotifierConfig interface {
	// IsEnabled 是否启用
	IsEnabled() bool
}

// NotificationStatus 通知状态
type NotificationStatus string

const (
	StatusPending NotificationStatus = "pending"
	StatusSuccess NotificationStatus = "success"
	StatusFailed  NotificationStatus = "failed"
)

// NotificationResult 通知结果
type NotificationResult struct {
	Channel      string             // 通知渠道
	Status       NotificationStatus // 通知状态
	TotalCount   int                // 总消息数
	SuccessCount int                // 成功消息数
	Error        string             // 错误信息
	StartAt      time.Time          // 开始时间
	EndAt        time.Time          // 结束时间
}

// Notifier 通知器接口
type Notifier interface {
	// Name 返回通知器名称
	Name() string
	// IsEnabled 检查是否启用
	IsEnabled() bool
	// Send 发送通知
	Send(ctx context.Context, items []MessageItem) (*NotificationResult, error)
}

// NotifierFactory 通知器工厂函数类型
type NotifierFactory func(config NotifierConfig) (Notifier, error)

// RegisterNotifierFunc 注册通知器的函数类型
type RegisterNotifierFunc func(registry *NotifierRegistry)

// NotifierRegistry 通知器注册表
type NotifierRegistry struct {
	factories map[string]NotifierFactory
}

// NewNotifierRegistry 创建通知器注册表
func NewNotifierRegistry() *NotifierRegistry {
	return &NotifierRegistry{
		factories: make(map[string]NotifierFactory),
	}
}

// Register 注册通知器工厂
func (r *NotifierRegistry) Register(name string, factory NotifierFactory) {
	r.factories[name] = factory
}

// Get 获取通知器工厂
func (r *NotifierRegistry) Get(name string) (NotifierFactory, bool) {
	factory, exists := r.factories[name]
	return factory, exists
}

// List 获取所有注册的通知器名称
func (r *NotifierRegistry) List() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}
