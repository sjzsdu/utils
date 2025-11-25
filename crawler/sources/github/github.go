package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sjzsdu/utils/crawler/internal/fetcher"
	"github.com/sjzsdu/utils/crawler/internal/parser"
	"github.com/sjzsdu/utils/crawler/pkg/models"
	"github.com/sjzsdu/utils/crawler/sources"
)

// githubSource 是 GitHub 数据源的实现
type githubSource struct {
	name     string
	url      string
	interval int
	client   *fetcher.HttpClient
	parser   *parser.HTMLParser
}

// NewGitHubSource 创建一个新的 GitHub 数据源实例
func NewGitHubSource() *githubSource {
	return &githubSource{
		name:     "github",
		url:      "https://github.com/trending",
		interval: 3600, // 1小时
		client:   fetcher.NewHttpClient(10 * time.Second),
		parser:   parser.NewHTMLParser(),
	}
}

// GetName 返回数据源名称
func (g *githubSource) GetName() string {
	return g.name
}

// GetURL 返回数据源的基础 URL
func (g *githubSource) GetURL() string {
	return g.url
}

// Fetch 获取数据源内容
func (g *githubSource) Fetch(ctx context.Context) ([]byte, error) {
	return g.client.Get(ctx, g.url)
}

// Parse 解析获取到的内容，返回结构化数据
func (g *githubSource) Parse(content []byte) ([]models.Item, error) {
	var items []models.Item

	// 使用 goquery 解析 HTML
	selections, err := g.parser.FindAll(content, ".Box .Box-row")
	if err != nil {
		return nil, err
	}

	now := time.Now()

	for i, s := range selections {
		// 提取标题和URL
		titleSel := s.Find(".h3 a")
		title := titleSel.Text()
		url, _ := titleSel.Attr("href")
		fullURL := "https://github.com" + url

		// 提取描述
		desc := s.Find(".col-9.color-fg-muted.my-1.pr-4").Text()

		// 提取语言
		lang := s.Find(".repo-language-color").Next().Text()

		// 生成ID
		id := fmt.Sprintf("github-%d-%d", now.Unix(), i)

		// 创建Item
		item := models.Item{
			ID:          id,
			Title:       strings.TrimSpace(title),
			URL:         fullURL,
			Content:     strings.TrimSpace(desc),
			Source:      g.GetName(),
			Category:    strings.TrimSpace(lang),
			Images:      []string{},
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		items = append(items, item)
	}

	return items, nil
}

// GetInterval 返回爬取间隔（秒）
func (g *githubSource) GetInterval() int {
	return g.interval
}

func init() {
	// 注册GitHub数据源
	source := NewGitHubSource()
	sources.GetRegistry().Register(source)
}
