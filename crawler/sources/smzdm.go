package sources

import (
	"bytes"
	"context"
	"io"
	"net/http"
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

// Fetch 获取什么值得买热门内容
func (s *SmzdmSource) Fetch(ctx context.Context) ([]byte, error) {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置完整的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	client := s.Client
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
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
