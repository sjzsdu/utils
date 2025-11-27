package sources

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// SmzdmSource 什么值得买数据源
// 该数据源从什么值得买网站提取热门内容
type SmzdmSource struct {
	BaseSource
}

// NewSmzdmSource 创建什么值得买数据源实例
func NewSmzdmSource() *SmzdmSource {
	return &SmzdmSource{
		BaseSource: BaseSource{
			Name:     "smzdm",
			URL:      "https://post.smzdm.com/hot_1/",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析什么值得买热门内容
func (s *SmzdmSource) Parse(content []byte) ([]models.Item, error) {
	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	items := make([]models.Item, 0)

	// 查找所有热门标题
	doc.Find("#feed-main-list .z-feed-title").Each(func(i int, el *goquery.Selection) {
		a := el.Find("a")
		url, exists := a.Attr("href")
		if !exists {
			return
		}

		title := a.Text()
		if title == "" {
			return
		}

		items = append(items, models.Item{
			ID:          url,
			Title:       title,
			URL:         url,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	})

	return items, nil
}

func init() {
	RegisterSource(NewSmzdmSource())
}
