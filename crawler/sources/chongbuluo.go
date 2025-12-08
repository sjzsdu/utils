package sources

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// ChongbuluoSource 虫部落数据源
// 该数据源从虫部落网站提取热门内容
type ChongbuluoSource struct {
	BaseSource
}

// NewChongbuluoSource 创建虫部落数据源实例
func NewChongbuluoSource() *ChongbuluoSource {
	return &ChongbuluoSource{
		BaseSource: BaseSource{
			Name:       "chongbuluo",
			URL:        "https://www.chongbuluo.com/forum.php?mod=guide&view=hot",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技", "综合"},
		},
	}
}

// Parse 解析虫部落热门内容
func (s *ChongbuluoSource) Parse(content []byte) ([]models.Item, error) {
	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	items := make([]models.Item, 0)
	baseURL := "https://www.chongbuluo.com/"

	// 查找所有热门帖子
	doc.Find(".bmw table tr").Each(func(i int, el *goquery.Selection) {
		// 获取标题和链接
		title := el.Find(".common .xst").Text()
		url, exists := el.Find(".common a").Attr("href")
		if !exists || title == "" {
			return
		}

		items = append(items, models.Item{
			ID:          baseURL + url,
			Title:       title,
			URL:         baseURL + url,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	})

	return items, nil
}

func init() {
	RegisterSource(NewChongbuluoSource())
}
