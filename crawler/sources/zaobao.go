package sources

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ZaobaoSource 联合早报数据源
// 该数据源从联合早报获取实时新闻
type ZaobaoSource struct {
	BaseSource
}

// NewZaobaoSource 创建联合早报数据源实例
func NewZaobaoSource() *ZaobaoSource {
	return &ZaobaoSource{
		BaseSource: BaseSource{
			Name:     "zaobao",
			URL:      "https://www.zaochenbao.com/realtime/",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析联合早报新闻内容
func (s *ZaobaoSource) Parse(content []byte) ([]models.Item, error) {
	// 处理GB2312编码
	reader := transform.NewReader(bytes.NewReader(content), simplifiedchinese.GB18030.NewDecoder())
	decodedContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(decodedContent))
	if err != nil {
		return nil, err
	}

	var items []models.Item
	base := "https://www.zaochenbao.com"

	doc.Find("div.list-block>a.item").Each(func(i int, selection *goquery.Selection) {
		url, _ := selection.Attr("href")
		title := selection.Find(".eps").Text()
		dateStr := selection.Find(".pdt10").Text()

		if url != "" && title != "" && dateStr != "" {
			// 简单处理日期，这里可能需要更复杂的日期解析
			pubDate := time.Now()

			items = append(items, models.Item{
				ID:          url,
				Title:       title,
				URL:         base + url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: pubDate,
			})
		}
	})

	// 按发布时间排序
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].PublishedAt.Before(items[j].PublishedAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewZaobaoSource())
}
