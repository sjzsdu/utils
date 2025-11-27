package sources

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// Kr36Source 36氪数据源
type Kr36Source struct {
	BaseSource
}

// Kr36News 36氪新闻结构
type Kr36News struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Summary string    `json:"summary"`
	URL     string    `json:"url"`
	Time    time.Time `json:"time"`
}

// New36KrSource 创建36氪数据源实例
func New36KrSource() *Kr36Source {
	return &Kr36Source{
		BaseSource: BaseSource{
			Name:     "36kr",
			URL:      "https://36kr.com/api/newsflash/list",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析36氪新闻内容
func (s *Kr36Source) Parse(content []byte) ([]models.Item, error) {
	var resp struct {
		Data struct {
			Items []struct {
				ID          string    `json:"id"`
				Title       string    `json:"title"`
				Summary     string    `json:"summary"`
				NewsURL     string    `json:"news_url"`
				PublishedAt time.Time `json:"published_at"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	items := make([]models.Item, len(resp.Data.Items))
	for i, news := range resp.Data.Items {
		items[i] = models.Item{
			ID:        news.ID,
			Title:     news.Title,
			URL:       news.NewsURL,
			Content:   news.Summary,
			Source:    s.Name,
			CreatedAt: news.PublishedAt,
		}
	}

	return items, nil
}

// Fetch 获取36氪新闻内容
func (s *Kr36Source) Fetch(ctx context.Context) ([]byte, error) {
	// 36氪API可能需要特定的请求头，这里可以根据需要重写
	return s.BaseSource.Fetch(ctx)
}

func init() {
	RegisterSource(New36KrSource())
}
