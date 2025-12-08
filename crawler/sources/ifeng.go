package sources

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// IfengSource 凤凰网数据源
// 该数据源从凤凰网首页HTML中提取热点新闻数据
type IfengSource struct {
	BaseSource
}

// NewIfengSource 创建凤凰网数据源实例
func NewIfengSource() *IfengSource {
	return &IfengSource{
		BaseSource: BaseSource{
			Name:       "ifeng",
			URL:        "https://www.ifeng.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"综合", "时政"},
		},
	}
}

// Parse 解析凤凰网热点新闻内容
func (s *IfengSource) Parse(content []byte) ([]models.Item, error) {
	html := string(content)

	// 正则表达式匹配JavaScript变量
	re := regexp.MustCompile(`var\s+allData\s*=\s*(\{[\s\S]*?\});`)
	match := re.FindStringSubmatch(html)

	if match == nil || len(match) < 2 {
		return []models.Item{}, nil
	}

	// 解析JSON数据
	var allData struct {
		HotNews1 []struct {
			URL      string `json:"url"`
			Title    string `json:"title"`
			NewsTime string `json:"newsTime"`
		} `json:"hotNews1"`
	}

	if err := json.Unmarshal([]byte(match[1]), &allData); err != nil {
		return nil, err
	}

	items := make([]models.Item, len(allData.HotNews1))
	for i, hotNews := range allData.HotNews1 {
		// 尝试解析发布时间
		var pubTime time.Time
		if hotNews.NewsTime != "" {
			// 假设时间格式为：2023-12-25 15:30:20
			pubTime, _ = time.Parse("2006-01-02 15:04:05", hotNews.NewsTime)
		} else {
			pubTime = time.Now()
		}

		items[i] = models.Item{
			ID:          hotNews.URL,
			Title:       hotNews.Title,
			URL:         hotNews.URL,
			Source:      s.Name,
			CreatedAt:   pubTime,
			PublishedAt: pubTime,
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewIfengSource())
}
