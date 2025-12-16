package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// BingSearch 实现bing搜索引擎
type BingSearch struct {
	apiKey  string
	timeout int
	headers map[string]string
}

// NewBingSearch 创建bing搜索引擎实例
func NewBingSearch(apiKey string, opts ...SearchOption) *BingSearch {
	// 如果未提供API密钥，从环境变量获取
	if apiKey == "" {
		apiKey = os.Getenv("BING_API_KEY")
	}

	cfg := &SearchConfig{
		APIKey:  apiKey,
		Timeout: 15, // 默认15秒超时
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &BingSearch{
		apiKey:  cfg.APIKey,
		timeout: cfg.Timeout,
		headers: cfg.Headers,
	}
}

// Name 返回搜索引擎名称
func (b *BingSearch) Name() string {
	return "bing"
}

// Search 执行搜索并返回结果
func (b *BingSearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// 构建API URL
	searchURL := fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=%s&count=%d",
		url.QueryEscape(query), limit)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(b.timeout) * time.Second,
	}

	// 发送GET请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Ocp-Apim-Subscription-Key", b.apiKey)

	// 添加自定义请求头
	for k, v := range b.headers {
		httpReq.Header.Set(k, v)
	}

	// 执行请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析结果
	results, err := parseBingSearchResults(body, limit)
	if err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %v", err)
	}

	return results, nil
}

// parseBingSearchResults 解析Bing搜索结果
func parseBingSearchResults(data []byte, limit int) ([]SearchResult, error) {
	var response struct {
		WebPages struct {
			Value []struct {
				Name    string `json:"name"`
				URL     string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"value"`
		} `json:"webPages"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(response.WebPages.Value))
	for i, item := range response.WebPages.Value {
		if i >= limit {
			break
		}
		results = append(results, SearchResult{
			Title:   item.Name,
			URL:     item.URL,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}
