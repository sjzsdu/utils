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

// GoogleSearch 实现google搜索引擎
type GoogleSearch struct {
	apiKey         string
	searchEngineId string
	timeout        int
	headers        map[string]string
}

// NewGoogleSearch 创建google搜索引擎实例
func NewGoogleSearch(apiKey, searchEngineId string, opts ...SearchOption) *GoogleSearch {
	// 如果未提供API密钥，从环境变量获取
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	// 如果未提供搜索引擎ID，从环境变量获取
	if searchEngineId == "" {
		searchEngineId = os.Getenv("GOOGLE_CSE_ID")
	}

	cfg := &SearchConfig{
		APIKey:  apiKey,
		Timeout: 15, // 默认15秒超时
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &GoogleSearch{
		apiKey:         apiKey,
		searchEngineId: searchEngineId,
		timeout:        cfg.Timeout,
		headers:        cfg.Headers,
	}
}

// Name 返回搜索引擎名称
func (g *GoogleSearch) Name() string {
	return "google"
}

// Search 执行搜索并返回结果
func (g *GoogleSearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// 构建API URL
	searchURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&num=%d",
		g.apiKey, g.searchEngineId, url.QueryEscape(query), limit)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(g.timeout) * time.Second,
	}

	// 发送GET请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	// 添加自定义请求头
	for k, v := range g.headers {
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
	results, err := parseGoogleSearchResults(body, limit)
	if err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %v", err)
	}

	return results, nil
}

// parseGoogleSearchResults 解析Google搜索结果
func parseGoogleSearchResults(data []byte, limit int) ([]SearchResult, error) {
	var response struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(response.Items))
	for i, item := range response.Items {
		if i >= limit {
			break
		}
		results = append(results, SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}
