package sources

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// ITHomeSource IT之家数据源
type ITHomeSource struct {
	BaseSource
}

// NewITHomeSource 创建IT之家数据源实例
func NewITHomeSource() *ITHomeSource {
	return &ITHomeSource{
		BaseSource: BaseSource{
			Name:       "ithome",
			URL:        "https://www.ithome.com/rss/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技"},
		},
	}
}

// Parse 解析IT之家新闻列表
func (s *ITHomeSource) Parse(content []byte) ([]models.Item, error) {
	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}

	items := make([]models.Item, 0)
	now := time.Now()
	sourceName := s.Name

	// 查找新闻列表项
	doc.Find("#list > div.fl > ul > li").Each(func(i int, sel *goquery.Selection) {
		// 查找链接和标题
		aTag := sel.Find("a.t")
		url, exists := aTag.Attr("href")
		if !exists {
			return
		}

		title := strings.TrimSpace(aTag.Text())
		if title == "" {
			return
		}

		// 查找日期
		date := strings.TrimSpace(sel.Find("i").Text())
		if date == "" {
			return
		}

		// 过滤广告
		isAd := strings.Contains(url, "lapin") ||
			strings.Contains(title, "神券") ||
			strings.Contains(title, "优惠") ||
			strings.Contains(title, "补贴") ||
			strings.Contains(title, "京东")

		if !isAd {
			items = append(items, models.Item{
				ID:          url,
				Title:       title,
				URL:         url,
				Content:     date,
				Source:      sourceName,
				Category:    "tech",
				PublishedAt: now,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
	})

	return items, nil
}

func init() {
	RegisterSource(NewITHomeSource())
}
