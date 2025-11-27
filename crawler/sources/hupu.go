package sources

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// HupuSource 虎扑数据源
// 该数据源从虎扑网站的HTML页面中提取热榜内容
type HupuSource struct {
	BaseSource
}

// NewHupuSource 创建虎扑数据源实例
func NewHupuSource() *HupuSource {
	return &HupuSource{
		BaseSource: BaseSource{
			Name:     "hupu",
			URL:      "https://bbs.hupu.com/topic-daily-hot",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析虎扑热榜内容
func (s *HupuSource) Parse(content []byte) ([]models.Item, error) {
	html := string(content)

	// 正则表达式匹配热榜项结构
	re := regexp.MustCompile(`<li class="bbs-sl-web-post-body">[\s\S]*?<a href="(\/[^\"]+?\.html)"[^>]*?class="p-title"[^>]*>([^<]+)<\/a>`)
	matches := re.FindAllStringSubmatch(html, -1)

	if matches == nil {
		return []models.Item{}, nil
	}

	items := make([]models.Item, 0, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		path := match[1]
		title := strings.TrimSpace(match[2])
		url := fmt.Sprintf("https://bbs.hupu.com%s", path)

		items = append(items, models.Item{
			ID:          path,
			Title:       title,
			URL:         url,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewHupuSource())
}
