package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// SSPaiSource 少数派数据源
type SSPaiSource struct {
	BaseSource
}

// SSPaiResult 少数派API响应结构
type SSPaiResult struct {
	Data []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"data"`
}

// NewSSPaiSource 创建少数派数据源实例
func NewSSPaiSource() *SSPaiSource {
	return &SSPaiSource{
		BaseSource: BaseSource{
			Name:       "sspai",
			URL:        "https://sspai.com/api/v1/article/tag/page/get?limit=30&offset=0&created_at={{timestamp}}&tag=%E7%83%AD%E9%97%A8%E6%96%87%E7%AB%A0&released=false",
			Interval:   3600, // 1小时爬取一次
			Categories: []string{"科技", "应用"},
		},
	}
}

// Fetch 获取少数派热门文章
func (s *SSPaiSource) Fetch(ctx context.Context) ([]byte, error) {
	// 替换URL中的时间戳占位符
	url := strings.Replace(s.URL, "{{timestamp}}", fmt.Sprintf("%d", time.Now().UnixMilli()), 1)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")

	// 获取客户端，如果为nil则创建默认客户端
	client := s.Client
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch: %d", resp.StatusCode)
	}

	// 读取响应内容
	return io.ReadAll(resp.Body)
}

// Parse 解析少数派热门文章
func (s *SSPaiSource) Parse(content []byte) ([]models.Item, error) {
	var result SSPaiResult
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	items := make([]models.Item, 0, len(result.Data))
	now := time.Now()
	for _, item := range result.Data {
		items = append(items, models.Item{
			ID:          fmt.Sprintf("%d", item.ID),
			Title:       item.Title,
			URL:         fmt.Sprintf("https://sspai.com/post/%d", item.ID),
			Source:      s.Name,
			Category:    "tech",
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewSSPaiSource())
}
