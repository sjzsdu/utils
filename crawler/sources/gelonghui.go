package sources

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// GelonghuiSource 格隆汇数据源
// 该数据源从格隆汇获取财经新闻
type GelonghuiSource struct {
	BaseSource
}

// NewGelonghuiSource 创建格隆汇数据源实例
func NewGelonghuiSource() *GelonghuiSource {
	return &GelonghuiSource{
		BaseSource: BaseSource{
			Name:     "gelonghui",
			URL:      "https://www.gelonghui.com/news/",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析格隆汇新闻内容
func (s *GelonghuiSource) Parse(content []byte) ([]models.Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	var items []models.Item
	baseURL := "https://www.gelonghui.com"

	doc.Find(".article-content").Each(func(i int, selection *goquery.Selection) {
		a := selection.Find(".detail-right>a")
			url, _ := a.Attr("href")
			title := a.Find("h2").Text()
			relativeTime := selection.Find(".time > span:nth-child(3)").Text()

		if url != "" && title != "" && relativeTime != "" {
			// 简单处理时间，实际项目中可能需要更复杂的日期解析
			pubDate := time.Now()

			items = append(items, models.Item{
				ID:          url,
				Title:       title,
				URL:         baseURL + url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: pubDate,
			})
		}
	})

	return items, nil
}

func init() {
	RegisterSource(NewGelonghuiSource())
}
