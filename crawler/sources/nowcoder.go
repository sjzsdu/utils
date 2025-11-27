package sources

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// NowcoderSource 牛客网数据源
// 该数据源从牛客网获取热门搜索内容
type NowcoderSource struct {
	BaseSource
}

// NowcoderResult 牛客网结果项
type NowcoderResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  int    `json:"type"`
	UUID  string `json:"uuid"`
}

// NowcoderResponse 牛客网响应
type NowcoderResponse struct {
	Data struct {
		Result []NowcoderResult `json:"result"`
	} `json:"data"`
}

// NewNowcoderSource 创建牛客网数据源实例
func NewNowcoderSource() *NowcoderSource {
	return &NowcoderSource{
		BaseSource: BaseSource{
			Name:     "nowcoder",
			URL:      "https://gw-c.nowcoder.com/api/sparta/hot-search/top-hot-pc",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Fetch 获取牛客网热门搜索数据
func (s *NowcoderSource) Fetch(ctx context.Context) ([]byte, error) {
	// 添加时间戳参数避免缓存
	timestamp := time.Now().UnixMilli()
	originalURL := s.URL
	s.URL = originalURL + "?size=20&_=" + strconv.FormatInt(timestamp, 10) + "&t="
	defer func() {
		s.URL = originalURL
	}()

	return s.BaseSource.Fetch(ctx)
}

// Parse 解析牛客网热门搜索内容
func (s *NowcoderSource) Parse(content []byte) ([]models.Item, error) {
	var resp NowcoderResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item
	for _, result := range resp.Data.Result {
		var url, id string

		if result.Type == 74 {
			url = "https://www.nowcoder.com/feed/main/detail/" + result.UUID
			id = result.UUID
		} else if result.Type == 0 {
			url = "https://www.nowcoder.com/discuss/" + result.ID
			id = result.ID
		} else {
			// 跳过其他类型
			continue
		}

		if url != "" && id != "" && result.Title != "" {
			items = append(items, models.Item{
				ID:          id,
				Title:       result.Title,
				URL:         url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Now(),
			})
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewNowcoderSource())
}
