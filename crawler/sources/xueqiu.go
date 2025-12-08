package sources

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// XueqiuSource 雪球数据源
// 该数据源从雪球网站获取热门股票数据
type XueqiuSource struct {
	BaseSource
}

// XueqiuStockItem 雪球股票项
type XueqiuStockItem struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Percent  float64 `json:"percent"`
	Exchange string  `json:"exchange"`
	AD       int     `json:"ad"` // 1表示广告
}

// XueqiuResponse 雪球响应
type XueqiuResponse struct {
	Data struct {
		Items []XueqiuStockItem `json:"items"`
	} `json:"data"`
}

// NewXueqiuSource 创建雪球数据源实例
func NewXueqiuSource() *XueqiuSource {
	return &XueqiuSource{
		BaseSource: BaseSource{
			Name:       "xueqiu",
			URL:        "https://xueqiu.com/statuses/hot/listV2.json?since_id=-1&max_id=-1&size=15",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"财经"},
		},
	}
}

// Fetch 获取雪球热门股票数据
func (s *XueqiuSource) Fetch(ctx context.Context) ([]byte, error) {
	// 首先获取cookie
	hqURL := "https://xueqiu.com/hq"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hqURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 发送请求获取cookie
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 提取cookie
	var cookies []string
	for _, cookie := range resp.Cookies() {
		cookies = append(cookies, cookie.String())
	}
	cookieStr := strings.Join(cookies, "; ")

	// 然后使用cookie请求热门股票数据
	stockURL := "https://stock.xueqiu.com/v5/stock/hot_stock/list.json?size=30&_type=10&type=10"
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, stockURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req2.Header.Set("Cookie", cookieStr)

	// 发送请求获取股票数据
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	return io.ReadAll(resp2.Body)
}

// Parse 解析雪球热门股票内容
func (s *XueqiuSource) Parse(content []byte) ([]models.Item, error) {
	var resp XueqiuResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	items := make([]models.Item, 0)

	// 过滤掉广告项
	for _, stock := range resp.Data.Items {
		if stock.AD == 1 {
			continue // 跳过广告
		}

		items = append(items, models.Item{
			ID:          stock.Code,
			Title:       stock.Name,
			URL:         `https://xueqiu.com/s/` + stock.Code,
			Content:     stock.Code,
			Source:      s.Name,
			Category:    `股票`,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewXueqiuSource())
}
