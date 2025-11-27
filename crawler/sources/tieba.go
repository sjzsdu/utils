package sources

import (
	"encoding/json"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// TiebaSource 百度贴吧数据源
// 该数据源从百度贴吧热门话题列表获取数据
type TiebaSource struct {
	BaseSource
}

// TiebaTopic 百度贴吧话题项
type TiebaTopic struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	MobileURL string `json:"mobileUrl"`
}

// TiebaResponse 百度贴吧响应
type TiebaResponse struct {
	TopicList []TiebaTopic `json:"topic_list"`
}

// NewTiebaSource 创建百度贴吧数据源实例
func NewTiebaSource() *TiebaSource {
	return &TiebaSource{
		BaseSource: BaseSource{
			Name:     "tieba",
			URL:      "https://tieba.baidu.com/hottopic/browse/topicList",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析百度贴吧热门话题内容
func (s *TiebaSource) Parse(content []byte) ([]models.Item, error) {
	var resp struct {
		Data TiebaResponse `json:"data"`
	}
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	items := make([]models.Item, len(resp.Data.TopicList))
	for i, topic := range resp.Data.TopicList {
		items[i] = models.Item{
			ID:          topic.ID,
			Title:       topic.Title,
			URL:         topic.URL,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewTiebaSource())
}
