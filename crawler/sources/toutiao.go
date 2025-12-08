package sources

import (
	"encoding/json"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// ToutiaoSource 头条数据源
type ToutiaoSource struct {
	BaseSource
}

// ToutiaoImage 头条图片结构
type ToutiaoImage struct {
	URL string `json:"url"`
}

// ToutiaoLabelUri 头条标签URI
type ToutiaoLabelUri struct {
	URL string `json:"url"`
}

// ToutiaoItem 头条数据项
type ToutiaoItem struct {
	ClusterIdStr string        `json:"ClusterIdStr"`
	Title        string        `json:"Title"`
	HotValue     string        `json:"HotValue"`
	Image        ToutiaoImage  `json:"Image"`
	LabelUri     *ToutiaoLabelUri `json:"LabelUri,omitempty"`
}

// ToutiaoResponse 头条响应
type ToutiaoResponse struct {
	Data []ToutiaoItem `json:"data"`
}

// NewToutiaoSource 创建头条数据源实例
func NewToutiaoSource() *ToutiaoSource {
	return &ToutiaoSource{
		BaseSource: BaseSource{
			Name:       "toutiao",
			URL:        "https://www.toutiao.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"综合", "时政", "娱乐"},
		},
	}
}

// Parse 解析头条热点内容
func (s *ToutiaoSource) Parse(content []byte) ([]models.Item, error) {
	var resp ToutiaoResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item

	for _, item := range resp.Data {
		items = append(items, models.Item{
			ID:          item.ClusterIdStr,
			Title:       item.Title,
			URL:         `https://www.toutiao.com/trending/` + item.ClusterIdStr + `/`,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewToutiaoSource())
}
