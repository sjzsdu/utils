package sources

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// ZhihuSource 知乎数据源
type ZhihuSource struct {
	BaseSource
}

// ZhihuTitleArea 知乎标题区域
type ZhihuTitleArea struct {
	Text string `json:"text"`
}

// ZhihuExcerptArea 知乎摘要区域
type ZhihuExcerptArea struct {
	Text string `json:"text"`
}

// ZhihuImageArea 知乎图片区域
type ZhihuImageArea struct {
	URL string `json:"url"`
}

// ZhihuMetricsArea 知乎 metrics 区域
type ZhihuMetricsArea struct {
	Text       string `json:"text"`
	FontColor  string `json:"font_color"`
	Background string `json:"background"`
	Weight     string `json:"weight"`
}

// ZhihuLabelArea 知乎标签区域
type ZhihuLabelArea struct {
	Type        string `json:"type"`
	Trend       int64  `json:"trend"`
	NightColor  string `json:"night_color"`
	NormalColor string `json:"normal_color"`
}

// ZhihuLink 知乎链接
type ZhihuLink struct {
	URL string `json:"url"`
}

// ZhihuTarget 知乎目标对象
type ZhihuTarget struct {
	TitleArea   ZhihuTitleArea   `json:"title_area"`
	ExcerptArea ZhihuExcerptArea `json:"excerpt_area"`
	ImageArea   ZhihuImageArea   `json:"image_area"`
	MetricsArea ZhihuMetricsArea `json:"metrics_area"`
	LabelArea   ZhihuLabelArea   `json:"label_area"`
	Link        ZhihuLink        `json:"link"`
}

// ZhihuFeedSpecific 知乎 feed 特定字段
type ZhihuFeedSpecific struct {
	AnswerCount int64 `json:"answer_count"`
}

// ZhihuDataItem 知乎数据项
type ZhihuDataItem struct {
	Type         string            `json:"type"`
	StyleType    string            `json:"style_type"`
	FeedSpecific ZhihuFeedSpecific `json:"feed_specific"`
	Target       ZhihuTarget       `json:"target"`
}

// ZhihuResponse 知乎响应
type ZhihuResponse struct {
	Data []ZhihuDataItem `json:"data"`
}

// NewZhihuSource 创建知乎数据源实例
func NewZhihuSource() *ZhihuSource {
	return &ZhihuSource{
		BaseSource: BaseSource{
			Name:       "zhihu",
			URL:        "https://www.zhihu.com/api/v3/feed/topstory/hot-list-web?limit=20&desktop=true",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"综合", "时政"},
		},
	}
}

// Parse 解析知乎热点内容
func (s *ZhihuSource) Parse(content []byte) ([]models.Item, error) {
	var resp ZhihuResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item
	// 正则表达式匹配 URL 末尾的数字
	re := regexp.MustCompile(`(\d+)$`)

	for _, item := range resp.Data {
		// 从 URL 中提取 ID
		id := re.FindStringSubmatch(item.Target.Link.URL)
		if len(id) < 2 {
			// 如果没有匹配到数字，使用 URL 作为 ID
			id = []string{item.Target.Link.URL, item.Target.Link.URL}
		}

		items = append(items, models.Item{
			ID:          id[1],
			Title:       item.Target.TitleArea.Text,
			URL:         item.Target.Link.URL,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewZhihuSource())
}
