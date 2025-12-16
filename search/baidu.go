package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// BaiduSearch 实现baidu搜索引擎
type BaiduSearch struct {
	apiKey  string
	timeout int
	headers map[string]string
}

// NewBaiduSearch 创建baidu搜索引擎实例
func NewBaiduSearch(apiKey string, opts ...SearchOption) *BaiduSearch {
	// 如果未提供API密钥，从环境变量获取
	if apiKey == "" {
		apiKey = os.Getenv("BAIDU_API_KEY")
	}

	cfg := &SearchConfig{
		APIKey:  apiKey,
		Timeout: 30, // 默认30秒超时，百度API可能较慢
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &BaiduSearch{
		apiKey:  cfg.APIKey,
		timeout: cfg.Timeout,
		headers: cfg.Headers,
	}
}

// Name 返回搜索引擎名称
func (b *BaiduSearch) Name() string {
	return "baidu"
}

// Search 执行搜索并返回结果
func (b *BaiduSearch) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// 尝试百度千帆AI搜索API
	results, err := b.searchWithBaiduQianfanAPI(ctx, query, limit)
	if err != nil {
		// 如果千帆API失败，可以在这里添加备用的百度搜索API
		return nil, fmt.Errorf("百度搜索失败: %v", err)
	}
	return results, nil
}

// searchWithBaiduQianfanAPI 使用百度千帆AI搜索API
func (b *BaiduSearch) searchWithBaiduQianfanAPI(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// 构建API URL
	searchURL := "https://qianfan.baidubce.com/v2/ai_search/chat/completions"

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(b.timeout) * time.Second,
	}

	// 构建请求数据
	requestData := map[string]interface{}{
		"messages": []map[string]string{
			{
				"content": query,
				"role":    "user",
			},
		},
		"search_source": "baidu_search_v2",
		"resource_type_filter": []map[string]interface{}{
			{
				"type":  "web",
				"top_k": limit,
			},
		},
		"search_filter": map[string]interface{}{
			"match": map[string]interface{}{
				"site": []string{}, // 可以指定搜索特定网站
			},
		},
		"search_recency_filter": "year",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 发送POST请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+b.apiKey)

	// 添加自定义请求头
	for k, v := range b.headers {
		httpReq.Header.Set(k, v)
	}

	// 执行请求
	resp, err := client.Do(httpReq)
	if err != nil {
		// 更详细的错误处理
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, fmt.Errorf("请求百度API超时 (%d秒): %v", b.timeout, err)
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析结果
	results, err := parseBaiduAPISearchResults(body, limit)
	if err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %v", err)
	}

	return results, nil
}

// parseBaiduAPISearchResults 解析百度API搜索结果
func parseBaiduAPISearchResults(data []byte, limit int) ([]SearchResult, error) {
	// 先尝试通用的 JSON 解析
	var genericResponse map[string]interface{}
	if err := json.Unmarshal(data, &genericResponse); err != nil {
		return nil, fmt.Errorf("无法解析JSON响应: %v", err)
	}

	// 检查并解析千帆API的响应格式
	results := []SearchResult{}

	// 检查响应中是否有references字段 - 千帆AI搜索结构
	if refs, ok := genericResponse["references"].([]interface{}); ok {
		for i, ref := range refs {
			if i >= limit {
				break
			}
			if refMap, ok := ref.(map[string]interface{}); ok {
				title, _ := refMap["title"].(string)
				url, _ := refMap["url"].(string)
				content, _ := refMap["content"].(string)
				results = append(results, SearchResult{
					Title:   title,
					URL:     url,
					Snippet: content,
				})
			}
		}
		return results, nil
	}

	// 检查并解析其他可能的响应格式
	if search, ok := genericResponse["search"].(map[string]interface{}); ok {
		if web, ok := search["web"].([]interface{}); ok {
			for i, item := range web {
				if i >= limit {
					break
				}
				if itemMap, ok := item.(map[string]interface{}); ok {
					title, _ := itemMap["title"].(string)
					url, _ := itemMap["url"].(string)
					snippet, _ := itemMap["snippet"].(string)
					results = append(results, SearchResult{
						Title:   title,
						URL:     url,
						Snippet: snippet,
					})
				}
			}
			if len(results) > 0 {
				return results, nil
			}
		}
	}

	// 检查result.result路径
	if result, ok := genericResponse["result"].(map[string]interface{}); ok {
		if resultItems, ok := result["result"].([]interface{}); ok {
			for i, item := range resultItems {
				if i >= limit {
					break
				}
				if itemMap, ok := item.(map[string]interface{}); ok {
					title, _ := itemMap["title"].(string)
					url, _ := itemMap["url"].(string)
					snippet, _ := itemMap["snippet"].(string)
					results = append(results, SearchResult{
						Title:   title,
						URL:     url,
						Snippet: snippet,
					})
				}
			}
			if len(results) > 0 {
				return results, nil
			}
		}
	}

	// 检查错误信息
	if err, ok := genericResponse["error"].(map[string]interface{}); ok {
		message, _ := err["message"].(string)
		code, _ := err["code"].(string)
		if message != "" {
			return nil, fmt.Errorf("API错误: %s (代码: %s)", message, code)
		}
	}

	// 如果没有找到任何结果但也没有错误，返回空结果集
	return results, nil
}
