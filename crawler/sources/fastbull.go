package sources

import (
	"bytes"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// FastbullExpressSource 快牛快讯数据源
type FastbullExpressSource struct {
	BaseSource
}

// FastbullNewsSource 快牛新闻数据源
type FastbullNewsSource struct {
	BaseSource
}

// NewFastbullExpressSource 创建快牛快讯数据源实例
func NewFastbullExpressSource() *FastbullExpressSource {
	return &FastbullExpressSource{
		BaseSource: BaseSource{
			Name:       "fastbull-express",
			URL:        "https://www.fastbull.com/cn/express-news",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"财经", "科技"},
		},
	}
}

// NewFastbullNewsSource 创建快牛新闻数据源实例
func NewFastbullNewsSource() *FastbullNewsSource {
	return &FastbullNewsSource{
		BaseSource: BaseSource{
			Name:       "fastbull-news",
			URL:        "https://www.fastbull.com/cn/news",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"财经", "科技"},
		},
	}
}

// Parse 解析快牛快讯内容
func (s *FastbullExpressSource) Parse(content []byte) ([]models.Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	var items []models.Item
	baseURL := "https://www.fastbull.com"
	re := regexp.MustCompile(`【(.+)】`)

	doc.Find(".news-list").Each(func(i int, selection *goquery.Selection) {
		a := selection.Find(".title_name")
		url, _ := a.Attr("href")
		titleText := a.Text()
		dateStr, _ := selection.Attr("data-date")

		if url != "" && titleText != "" && dateStr != "" {
			// 提取标题中的内容
			title := titleText
			matches := re.FindStringSubmatch(titleText)
			if len(matches) > 1 {
				title = matches[1]
			}

			// 如果提取的标题太短，使用完整标题
			if len(title) < 4 {
				title = titleText
			}

			// 解析时间戳
			timestamp, err := strconv.ParseInt(dateStr, 10, 64)
			if err != nil {
				return
			}

			items = append(items, models.Item{
				ID:          url,
				Title:       title,
				URL:         baseURL + url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Unix(timestamp/1000, 0),
			})
		}
	})

	return items, nil
}

// Parse 解析快牛新闻内容
func (s *FastbullNewsSource) Parse(content []byte) ([]models.Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	var items []models.Item
	baseURL := "https://www.fastbull.com"

	doc.Find(".trending_type").Each(func(i int, selection *goquery.Selection) {
		url, _ := selection.Attr("href")
		title := selection.Find(".title").Text()
		dateStr, _ := selection.Find("[data-date]").Attr("data-date")

		if url != "" && title != "" && dateStr != "" {
			// 解析时间戳
			timestamp, err := strconv.ParseInt(dateStr, 10, 64)
			if err != nil {
				return
			}

			items = append(items, models.Item{
				ID:          url,
				Title:       title,
				URL:         baseURL + url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Unix(timestamp/1000, 0),
			})
		}
	})

	return items, nil
}

// FastbullSource 快牛主数据源
type FastbullSource struct {
	FastbullExpressSource
}

// NewFastbullSource 创建快牛主数据源实例
func NewFastbullSource() *FastbullSource {
	source := &FastbullSource{
		FastbullExpressSource: *NewFastbullExpressSource(),
	}
	source.Name = "fastbull"
	return source
}

func init() {
	// 注册三个数据源：主数据源、快讯数据源、新闻数据源
	RegisterSource(NewFastbullSource())
	RegisterSource(NewFastbullExpressSource())
	RegisterSource(NewFastbullNewsSource())
}
