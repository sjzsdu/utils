package sources

import (
	"bytes"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// SolidotSource Solidot数据源
// 该数据源从Solidot网站提取新闻内容
type SolidotSource struct {
	BaseSource
}

// NewSolidotSource 创建Solidot数据源实例
func NewSolidotSource() *SolidotSource {
	return &SolidotSource{
		BaseSource: BaseSource{
			Name:       "solidot",
			URL:        "https://www.solidot.org/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技"},
		},
	}
}

// Parse 解析Solidot新闻内容
func (s *SolidotSource) Parse(content []byte) ([]models.Item, error) {
	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	items := make([]models.Item, 0)
	baseURL := "https://www.solidot.org"

	// 查找所有新闻块
	doc.Find(".block_m").Each(func(i int, el *goquery.Selection) {
		// 获取链接和标题
		a := el.Find(".bg_htit a").Last()
		path, exists := a.Attr("href")
		if !exists {
			return
		}

		title := strings.TrimSpace(a.Text())
		if title == "" {
			return
		}

		// 获取发布时间
		timeStr := el.Find(".talk_time").Text()
		// 正则匹配时间格式：发表于2023年12月25日15时30分
		timeRegex := regexp.MustCompile(`发表于(.*?)分`)
		timeMatch := timeRegex.FindStringSubmatch(timeStr)
		if timeMatch == nil || len(timeMatch) < 2 {
			return
		}

		// 格式化时间字符串
		rawDate := timeMatch[1]
		// 替换为标准格式：2023-12-25 15:30
		formattedDate := strings.ReplaceAll(rawDate, "年", "-")
		formattedDate = strings.ReplaceAll(formattedDate, "月", "-")
		formattedDate = strings.ReplaceAll(formattedDate, "日", " ")
		formattedDate = strings.ReplaceAll(formattedDate, "时", ":")

		// 解析时间
		pubTime, err := time.Parse("2006-01-02 15:04", formattedDate)
		if err != nil {
			pubTime = time.Now()
		}

		items = append(items, models.Item{
			ID:          path,
			Title:       title,
			URL:         baseURL + path,
			Source:      s.Name,
			CreatedAt:   pubTime,
			PublishedAt: pubTime,
		})
	})

	return items, nil
}

func init() {
	RegisterSource(NewSolidotSource())
}
