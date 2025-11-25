package news

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/internal/fetcher"
	"github.com/sjzsdu/utils/crawler/internal/parser"
	"github.com/sjzsdu/utils/crawler/pkg/models"
	"github.com/sjzsdu/utils/crawler/sources"
)

// newsSource 是新闻数据源的实现
type newsSource struct {
	name     string
	url      string
	interval int
	client   *fetcher.HttpClient
	parser   *parser.HTMLParser
}

// NewNewsSource 创建一个新的新闻数据源实例
func NewNewsSource() *newsSource {
	return &newsSource{
		name:     "news",
		url:      "https://news.ycombinator.com/",
		interval: 600, // 10分钟
		client:   fetcher.NewHttpClient(10 * time.Second),
		parser:   parser.NewHTMLParser(),
	}
}

// GetName 返回数据源名称
func (n *newsSource) GetName() string {
	return n.name
}

// GetURL 返回数据源的基础 URL
func (n *newsSource) GetURL() string {
	return n.url
}

// Fetch 获取数据源内容
func (n *newsSource) Fetch(ctx context.Context) ([]byte, error) {
	return n.client.Get(ctx, n.url)
}

// Parse 解析获取到的内容，返回结构化数据
func (n *newsSource) Parse(content []byte) ([]models.Item, error) {
	var items []models.Item

	// 使用 goquery 解析 HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 提取新闻条目
	doc.Find(".athing").Each(func(i int, s *goquery.Selection) {
		// 提取排名
		rank := s.Find(".rank").Text()

		// 提取标题和URL
		titleSel := s.Find(".titleline a")
		title := titleSel.Text()
		url, _ := titleSel.Attr("href")

		// 提取下一行的点数和评论数
		sibling := s.Next()
		points := sibling.Find(".score").Text()

		// 生成ID
		id := fmt.Sprintf("news-%d-%d", now.Unix(), i)

		// 创建Item
		item := models.Item{
			ID:          id,
			Title:       fmt.Sprintf("%s. %s", rank, title),
			URL:         url,
			Content:     points,
			Source:      n.name,
			Category:    "technology",
			Images:      []string{},
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		items = append(items, item)
	})

	return items, nil
}

// GetInterval 返回爬取间隔（秒）
func (n *newsSource) GetInterval() int {
	return n.interval
}

func init() {
	// 注册新闻数据源
	source := NewNewsSource()
	sources.GetRegistry().Register(source)
}
