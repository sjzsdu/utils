package sources

import (
	"encoding/xml"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// HackerNewsSource HackerNews数据源
type HackerNewsSource struct {
	BaseSource
}

// HackerNewsRSS HackerNews RSS结构
type HackerNewsRSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Item []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

// NewHackerNewsSource 创建HackerNews数据源实例
func NewHackerNewsSource() *HackerNewsSource {
	return &HackerNewsSource{
		BaseSource: BaseSource{
			Name:     "hackernews",
			URL:      "https://news.ycombinator.com/rss",
			Interval: 3600, // 1小时爬取一次
		},
	}
}

// Parse 解析HackerNews RSS内容
func (s *HackerNewsSource) Parse(content []byte) ([]models.Item, error) {
	var rss HackerNewsRSS
	if err := xml.Unmarshal(content, &rss); err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]models.Item, 0, len(rss.Channel.Item))
	for _, item := range rss.Channel.Item {
		// 提取ID
		id := item.Link
		if len(id) > 0 {
			// 从URL中提取ID
			for i := len(id) - 1; i >= 0; i-- {
				if id[i] == '=' {
					id = id[i+1:]
					break
				}
			}
		}

		items = append(items, models.Item{
			ID:          id,
			Title:       item.Title,
			URL:         item.Link,
			Content:     item.Description,
			Source:      s.Name,
			Category:    "news",
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewHackerNewsSource())
}
