package sources

import (
	"encoding/json"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// BilibiliSource Bilibili数据源
type BilibiliSource struct {
	BaseSource
}

// NewBilibiliSource 创建B站数据源实例
func NewBilibiliSource() *BilibiliSource {
	return &BilibiliSource{
		BaseSource: BaseSource{
			Name:       "bilibili",
			URL:        "https://www.bilibili.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"综合", "娱乐"},
		},
	}
}

// Parse 解析Bilibili热搜内容
func (s *BilibiliSource) Parse(content []byte) ([]models.Item, error) {
	var data struct {
		List []struct {
			Keyword  string `json:"keyword"`
			ShowName string `json:"show_name"`
			URL      string `json:"goto_value"`
		} `json:"list"`
	}

	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	items := make([]models.Item, 0, len(data.List))
	for _, item := range data.List {
		// 构建完整URL
		var fullURL string
		if item.URL != "" {
			fullURL = item.URL
		} else {
			fullURL = "https://search.bilibili.com/all?keyword=" + item.Keyword
		}

		items = append(items, models.Item{
			ID:      item.Keyword,
			Title:   item.ShowName,
			URL:     fullURL,
			Content: "",
			Source:  s.Name,
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewBilibiliSource())
}
