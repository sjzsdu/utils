package sources

import (
	"encoding/json"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// V2EXSource V2EX数据源
type V2EXSource struct {
	BaseSource
}

// V2EXResult V2EX API响应结构
type V2EXResult struct {
	Topics []struct {
		ID      int    `json:"id"`
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	}
}

// NewV2EXSource 创建V2EX数据源实例
func NewV2EXSource() *V2EXSource {
	return &V2EXSource{
		BaseSource: BaseSource{
			Name:     "v2ex",
			URL:      "https://www.v2ex.com/api/topics/hot.json",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析V2EX热门话题
func (s *V2EXSource) Parse(content []byte) ([]models.Item, error) {
	var topics []struct {
		ID      int    `json:"id"`
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
		Replies int    `json:"replies"`
	}

	if err := json.Unmarshal(content, &topics); err != nil {
		return nil, err
	}

	items := make([]models.Item, 0, len(topics))
	now := time.Now()
	for _, topic := range topics {
		items = append(items, models.Item{
			ID:          string(rune(topic.ID)),
			Title:       topic.Title,
			URL:         topic.URL,
			Content:     topic.Content,
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
	RegisterSource(NewV2EXSource())
}
