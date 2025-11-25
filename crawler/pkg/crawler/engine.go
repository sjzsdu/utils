package crawler

import (
	"context"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// Engine 定义了爬取引擎的基本行为
type Engine interface {
	// RegisterSource 注册数据源
	RegisterSource(source Source) error

	// UnregisterSource 注销数据源
	UnregisterSource(name string) error

	// Start 启动爬取引擎
	Start(ctx context.Context) error

	// Stop 停止爬取引擎
	Stop() error

	// FetchItem 获取指定数据源的最新数据
	FetchItem(ctx context.Context, sourceName string) ([]models.Item, error)

	// Subscribe 订阅数据源的更新
	Subscribe(sourceName string, ch chan<- []models.Item) error

	// Unsubscribe 取消订阅
	Unsubscribe(sourceName string, ch chan<- []models.Item) error
}
