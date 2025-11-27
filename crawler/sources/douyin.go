package sources

import (
	"context"
	"encoding/json"
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
			Name:     "douyin",
			URL:      "https://www.douyin.com/aweme/v1/web/hot/search/list/?device_platform=webapp&aid=6383&channel=channel_pc_web&detail_list=1",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Fetch 获取抖音热门搜索数据
func (s *DouyinSource) Fetch(ctx context.Context) ([]byte, error) {
	// 首先获取cookie
	cookieURL := "https://www.douyin.com/passport/general/login_guiding_strategy/?aid=6383"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cookieURL, nil)
	if err != nil {
		return nil, err
	}

	// 发送请求获取cookie
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 收集cookie
	var cookies []string
	for _, cookie := range resp.Cookies() {
		cookies = append(cookies, cookie.Name+"="+cookie.Value)
	}

	// 现在使用收集到的cookie请求热点数据
	hotReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置cookie头
	if len(cookies) > 0 {
		cookieStr := cookies[0]
		for i := 1; i < len(cookies); i++ {
			cookieStr += "; " + cookies[i]
		}
		hotReq.Header.Set("Cookie", cookieStr)
	}

	// 发送请求获取热点数据
	hotResp, err := client.Do(hotReq)
	if err != nil {
		return nil, err
	}
	defer hotResp.Body.Close()

	return readResponseBody(hotResp)
}

// Parse 解析抖音热门搜索内容
func (s *DouyinSource) Parse(content []byte) ([]models.Item, error) {
	var resp DouyinResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	items := make([]models.Item, len(resp.Data.WordList))
	for i, word := range resp.Data.WordList {
		items[i] = models.Item{
			ID:          word.SentenceID,
			Title:       word.Word,
			URL:         "https://www.douyin.com/hot/" + word.SentenceID,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewDouyinSource())
}
