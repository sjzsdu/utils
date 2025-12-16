package search

import (
	"context"
	"fmt"
)

// Client 定义搜索客户端
type Client struct {
	engines       map[string]SearchEngine
	defaultEngine string
}

// NewClient 创建搜索客户端实例
func NewClient() *Client {
	return &Client{
		engines: make(map[string]SearchEngine),
	}
}

// RegisterEngine 注册搜索引擎
func (c *Client) RegisterEngine(engine SearchEngine) {
	c.engines[engine.Name()] = engine
}

// SetDefaultEngine 设置默认搜索引擎
func (c *Client) SetDefaultEngine(name string) error {
	if _, ok := c.engines[name]; !ok {
		return fmt.Errorf("搜索引擎 %s 未注册", name)
	}
	c.defaultEngine = name
	return nil
}

// Search 执行搜索
func (c *Client) Search(ctx context.Context, query string, limit int, opts ...SearchOption) ([]SearchResult, error) {
	cfg := &SearchConfig{
		Engine: c.defaultEngine,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// 如果没有指定搜索引擎且没有默认搜索引擎，返回错误
	if cfg.Engine == "" {
		return nil, fmt.Errorf("未指定搜索引擎且没有设置默认搜索引擎")
	}

	// 获取搜索引擎
	engine, ok := c.engines[cfg.Engine]
	if !ok {
		return nil, fmt.Errorf("搜索引擎 %s 未注册", cfg.Engine)
	}

	// 执行搜索
	return engine.Search(ctx, query, limit)
}

// SearchWithEngine 指定搜索引擎执行搜索
func (c *Client) SearchWithEngine(ctx context.Context, engineName, query string, limit int) ([]SearchResult, error) {
	// 获取搜索引擎
	engine, ok := c.engines[engineName]
	if !ok {
		return nil, fmt.Errorf("搜索引擎 %s 未注册", engineName)
	}

	// 执行搜索
	return engine.Search(ctx, query, limit)
}

// ListEngines 返回已注册的搜索引擎列表
func (c *Client) ListEngines() []string {
	engines := make([]string, 0, len(c.engines))
	for name := range c.engines {
		engines = append(engines, name)
	}
	return engines
}

// NewDefaultClient 创建默认配置的搜索客户端
// 包含所有支持的搜索引擎
// 如果未提供API密钥，将尝试从环境变量获取
func NewDefaultClient(bingAPIKey, baiduAPIKey, googleAPIKey, googleSearchEngineID string, opts ...SearchOption) (*Client, error) {
	client := NewClient()

	// 注册bing搜索引擎（即使API密钥为空，也会从BING_API_KEY环境变量获取）
	bing := NewBingSearch(bingAPIKey, opts...)
	client.RegisterEngine(bing)

	// 注册baidu搜索引擎（即使API密钥为空，也会从BAIDU_API_KEY环境变量获取）
	baidu := NewBaiduSearch(baiduAPIKey, opts...)
	client.RegisterEngine(baidu)

	// 注册google搜索引擎（即使API密钥为空，也会从GOOGLE_API_KEY和GOOGLE_CSE_ID环境变量获取）
	google := NewGoogleSearch(googleAPIKey, googleSearchEngineID, opts...)
	client.RegisterEngine(google)

	// 如果没有注册任何搜索引擎，返回错误
	if len(client.engines) == 0 {
		return nil, fmt.Errorf("未注册任何搜索引擎")
	}

	// 默认使用第一个注册的搜索引擎
	for name := range client.engines {
		client.defaultEngine = name
		break
	}

	return client, nil
}
