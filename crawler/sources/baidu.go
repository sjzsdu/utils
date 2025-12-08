package sources

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// BaiduSource 百度数据源
type BaiduSource struct {
	BaseSource
}

// NewBaiduSource 创建百度数据源实例
func NewBaiduSource() *BaiduSource {
	return &BaiduSource{
		BaseSource: BaseSource{
			Name:       "baidu",
			URL:        "https://www.baidu.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"综合", "搜索"},
		},
	}
}

// Parse 解析百度热搜内容
func (s *BaiduSource) Parse(content []byte) ([]models.Item, error) {
	// 提取JSON数据
	contentStr := string(content)
	startIdx := strings.Index(contentStr, "<!--s-data:")
	endIdx := strings.Index(contentStr, "-->")
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("failed to find JSON data")
	}

	jsonStr := contentStr[startIdx+11 : endIdx]

	var data struct {
		Data struct {
			Cards []struct {
				Content []struct {
					Word   string `json:"word"`
					RawURL string `json:"rawUrl"`
					Desc   string `json:"desc"`
					IsTop  bool   `json:"isTop"`
				} `json:"content"`
			} `json:"cards"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	if len(data.Data.Cards) == 0 {
		return []models.Item{}, nil
	}

	items := make([]models.Item, 0, len(data.Data.Cards[0].Content))
	for _, item := range data.Data.Cards[0].Content {
		if item.IsTop {
			continue // 跳过置顶项
		}

		items = append(items, models.Item{
			ID:      item.RawURL,
			Title:   item.Word,
			URL:     item.RawURL,
			Content: item.Desc,
			Source:  s.Name,
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewBaiduSource())
}
