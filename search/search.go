package search

import (
	"context"
)

// SearchResult 表示搜索结果
type SearchResult struct {
	Title   string `json:"title"`   // 搜索结果标题
	URL     string `json:"url"`     // 搜索结果URL
	Snippet string `json:"snippet"` // 搜索结果摘要
}

// SearchEngine 定义搜索引擎接口
type SearchEngine interface {
	// Search 执行搜索并返回结果
	// ctx: 上下文，用于控制超时和取消
	// query: 搜索查询字符串
	// limit: 返回结果数量限制
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	// Name 返回搜索引擎名称
	Name() string
}

// SearchOption 定义搜索选项
type SearchOption func(*SearchConfig)

// SearchConfig 定义搜索配置
type SearchConfig struct {
	Engine   string            // 搜索引擎名称
	APIKey   string            // API密钥
	OtherKey string            // 其他密钥（如Google的Search Engine ID）
	Timeout  int               // 超时时间（秒）
	Headers  map[string]string // 自定义请求头
}

// WithEngine 设置搜索引擎
func WithEngine(engine string) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.Engine = engine
	}
}

// WithAPIKey 设置API密钥
func WithAPIKey(apiKey string) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.APIKey = apiKey
	}
}

// WithOtherKey 设置其他密钥
func WithOtherKey(otherKey string) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.OtherKey = otherKey
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout int) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.Timeout = timeout
	}
}

// WithHeaders 设置自定义请求头
func WithHeaders(headers map[string]string) SearchOption {
	return func(cfg *SearchConfig) {
		cfg.Headers = headers
	}
}
