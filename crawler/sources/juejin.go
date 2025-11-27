package sources

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// JuejinSource 掘金数据源
type JuejinSource struct {
	BaseSource
}

// NewJuejinSource 创建掘金数据源实例
func NewJuejinSource() *JuejinSource {
	return &JuejinSource{
		BaseSource: BaseSource{
			Name:     "juejin",
			URL:      "https://api.juejin.cn/content_api/v1/content/article_rank?category_id=1&type=hot&spider=0",
			Interval: 3600, // 1小时爬取一次
		},
	}
}

// JuejinResult 掘金API响应结构
type JuejinResult struct {
	Data []struct {
		Content struct {
			ContentID    string `json:"content_id"`
			Title        string `json:"title"`
			BriefContent string `json:"brief_content"`
			Category     struct {
				Name string `json:"name"`
			} `json:"category"`
		} `json:"content"`
	} `json:"data"`
}

// Parse 解析掘金热榜内容
func (s *JuejinSource) Parse(content []byte) ([]models.Item, error) {
	var result JuejinResult
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	items := make([]models.Item, 0, len(result.Data))
	now := time.Now()
	for _, item := range result.Data {
		items = append(items, models.Item{
			ID:          item.Content.ContentID,
			Title:       item.Content.Title,
			URL:         fmt.Sprintf("https://juejin.cn/post/%s", item.Content.ContentID),
			Content:     item.Content.BriefContent,
			Source:      s.Name,
			Category:    item.Content.Category.Name,
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewJuejinSource())
}
