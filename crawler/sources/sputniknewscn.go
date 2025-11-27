package sources

import (
	"bytes"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// SputniknewsCNSource 俄罗斯卫星通讯社中文网数据源
// 该数据源从俄罗斯卫星通讯社中文网获取最新新闻
type SputniknewsCNSource struct {
	BaseSource
}

// NewSputniknewsCNSource 创建俄罗斯卫星通讯社中文网数据源实例
func NewSputniknewsCNSource() *SputniknewsCNSource {
	return &SputniknewsCNSource{
		BaseSource: BaseSource{
			Name:     "sputniknewscn",
			URL:      "https://sputniknews.cn/services/widget/lenta/",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析俄罗斯卫星通讯社中文网新闻内容
func (s *SputniknewsCNSource) Parse(content []byte) ([]models.Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	var items []models.Item
	doc.Find(".lenta__item").Each(func(i int, selection *goquery.Selection) {
		link := selection.Find("a")
		url, _ := link.Attr("href")
		title := link.Find(".lenta__item-text").Text()
		dateStr, _ := link.Find(".lenta__item-date").Attr("data-unixtime")

		if url != "" && title != "" && dateStr != "" {
			// 将时间戳转换为时间
			timestamp, err := strconv.ParseInt(dateStr, 10, 64)
			if err != nil {
				return
			}

			items = append(items, models.Item{
				ID:          url,
				Title:       title,
				URL:         "https://sputniknews.cn" + url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Unix(timestamp, 0),
			})
		}
	})

	return items, nil
}

func init() {
	RegisterSource(NewSputniknewsCNSource())
}
