package sources

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// DouyinSource 抖音数据源
// 该数据源从抖音获取热门搜索内容
type DouyinSource struct {
	BaseSource
}

// DouyinWord 抖音热点词语
type DouyinWord struct {
	SentenceID string `json:"sentence_id"`
	Word       string `json:"word"`
	EventTime  string `json:"event_time"`
	HotValue   string `json:"hot_value"`
}

// DouyinResponse 抖音响应
type DouyinResponse struct {
	Data struct {
		WordList []DouyinWord `json:"word_list"`
	} `json:"data"`
}

// NewDouyinSource 创建抖音数据源实例
func NewDouyinSource() *DouyinSource {
	return &DouyinSource{
		BaseSource: BaseSource{
			Name:       "douyin",
			URL:        "https://www.douyin.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"娱乐", "综合"},
		},
	}
}

// Fetch 获取抖音热门搜索数据
func (s *DouyinSource) Fetch(ctx context.Context) ([]byte, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置查询参数
	q := req.URL.Query()
	q.Add("device_platform", "webapp")
	q.Add("aid", "6383")
	q.Add("channel", "channel_pc_web")
	req.URL.RawQuery = q.Encode()

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.douyin.com")
	req.Header.Set("Referer", "https://www.douyin.com/")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Parse 解析抖音热门搜索内容
func (s *DouyinSource) Parse(content []byte) ([]models.Item, error) {
	var resp DouyinResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		// 如果JSON解析失败，返回空数组而非错误，避免程序崩溃
		return []models.Item{}, nil
	}

	items := make([]models.Item, 0, len(resp.Data.WordList))
	for _, word := range resp.Data.WordList {
		if word.SentenceID == "" || word.Word == "" {
			continue
		}
		items = append(items, models.Item{
			ID:          word.SentenceID,
			Title:       word.Word,
			URL:         "https://www.douyin.com/hot/" + word.SentenceID,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewDouyinSource())
}
