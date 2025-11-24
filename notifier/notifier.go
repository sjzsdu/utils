package notifier

import (
	"context"
	"fmt"
	"sync"
)

// 注意：核心类型定义已移至types.go文件

// NotifierManager 通知管理器
type NotifierManager struct {
	notifiers []Notifier
	wg        sync.WaitGroup
}

// NewNotifierManager 创建通知管理器
func NewNotifierManager() (*NotifierManager, error) {
	manager := &NotifierManager{
		notifiers: make([]Notifier, 0),
	}

	return manager, nil
}

// RegisterNotifier 注册通知器
func (m *NotifierManager) RegisterNotifier(name string, notifier Notifier) {
	if notifier != nil && notifier.IsEnabled() {
		m.notifiers = append(m.notifiers, notifier)
	}
}

// SendToAll 发送到所有启用的通知渠道
func (m *NotifierManager) SendToAll(items []MessageItem) (map[string]*NotificationResult, error) {
	ctx := context.Background()
	// 如果没有通知渠道，直接返回空结果
	if len(m.notifiers) == 0 {
		return make(map[string]*NotificationResult), nil
	}

	results := make(map[string]*NotificationResult)
	errs := make([]error, 0)
	resultsChan := make(chan *NotificationResult, len(m.notifiers))
	errsChan := make(chan error, len(m.notifiers))

	// 并发发送通知
	for _, notifier := range m.notifiers {
		m.wg.Add(1)
		go func(n Notifier) {
			defer m.wg.Done()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				errsChan <- ctx.Err()
				return
			default:
				result, err := n.Send(ctx, items)
				if err != nil {
					errsChan <- fmt.Errorf("%s 发送失败: %w", n.Name(), err)
					return
				}
				resultsChan <- result
			}
		}(notifier)
	}

	// 等待所有通知发送完成
	m.wg.Wait()
	close(resultsChan)
	close(errsChan)

	// 收集结果
	for result := range resultsChan {
		results[result.Channel] = result
	}

	// 收集错误
	for err := range errsChan {
		errs = append(errs, err)
	}

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return results, errs[0]
	}

	return results, nil
}

// SendToSpecific 发送到指定的通知渠道
func (m *NotifierManager) SendToSpecific(channel string, items []MessageItem) (*NotificationResult, error) {
	ctx := context.Background()
	// 如果没有通知渠道，直接返回错误
	if len(m.notifiers) == 0 {
		return nil, fmt.Errorf("没有启用任何通知渠道")
	}

	for _, notifier := range m.notifiers {
		if notifier.Name() == channel && notifier.IsEnabled() {
			return notifier.Send(ctx, items)
		}
	}

	return nil, fmt.Errorf("通知渠道 %s 未启用或不存在", channel)
}

// GetEnabledChannels 获取已启用的通知渠道
func (m *NotifierManager) GetEnabledChannels() []string {
	channels := make([]string, 0, len(m.notifiers))
	for _, notifier := range m.notifiers {
		if notifier.IsEnabled() {
			channels = append(channels, notifier.Name())
		}
	}
	return channels
}
