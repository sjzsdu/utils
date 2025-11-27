package sources

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// KaopuItem 叩谱新闻项
type KaopuItem struct {
	Description string `json:"description"`
	Link        string `json:"link"`
	PubDate     string `json:"pub_date"`
	Publisher   string `json:"publisher"`
	Title       string `json:"title"`
}

// KaopuResponse 叩谱新闻响应
type KaopuResponse struct {
	Data []KaopuItem `json:"data"`
}

// KaopuSource 叩谱数据源
// 该数据源从Azure Blob存储获取新闻内容
type KaopuSource struct {
	BaseSource
}

// NewKaopuSource 创建叩谱数据源实例
func NewKaopuSource() *KaopuSource {
	return &KaopuSource{
		BaseSource: BaseSource{
			Name:     "kaopu",
			URL:      "https://kaopustorage.blob.core.windows.net/news-prod/news_list_hans_0.json",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析叩谱新闻内容
func (s *KaopuSource) Parse(content []byte) ([]models.Item, error) {
	// 直接将响应解析为数组类型
	var result []KaopuItem
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	var news []models.Item
	// 过滤掉特定出版商的内容
	for _, item := range result {
		// 排除"财新"和"公视"
		if item.Publisher != "财新" && item.Publisher != "公视" {
			// 解析发布日期
			pubDate, err := time.Parse(time.RFC3339, item.PubDate)
			if err != nil {
				pubDate = time.Now()
			}

			news = append(news, models.Item{
				ID:          fmt.Sprintf("%s", item.Link),
				Title:       item.Title,
				URL:         fmt.Sprintf("https://www.kaopu001.com%s", item.Link),
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: pubDate,
			})
		}
	}

	return news, nil
}

func init() {
	RegisterSource(NewKaopuSource())
}
