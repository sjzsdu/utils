package crawler

import (
	"context"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// Source 定义了数据源的基本行为
type Source interface {
	// GetName 返回数据源名称
	GetName() string

	// GetURL 返回数据源的基础 URL
	GetURL() string

	// Fetch 获取数据源内容
	Fetch(ctx context.Context) ([]byte, error)

	// Parse 解析获取到的内容，返回结构化数据
	Parse(content []byte) ([]models.Item, error)

	// GetInterval 返回爬取间隔（秒）
	GetInterval() int

	// GetCategories 返回数据源的分类列表
	GetCategories() []string
}
