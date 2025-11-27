package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// CoolapkSource 酷安数据源
type CoolapkSource struct {
	BaseSource
}

// CoolapkTargetRow 酷安目标行数据
type CoolapkTargetRow struct {
	SubTitle string `json:"subTitle"`
}

// CoolapkDataItem 酷安数据项
type CoolapkDataItem struct {
	ID          string            `json:"id"`
	Message     string            `json:"message"`      // 多行
	EditorTitle string            `json:"editor_title"` // 起的标题
	URL         string            `json:"url"`
	EntityType  string            `json:"entityType"`
	PubDate     string            `json:"pubDate"`
	Dateline    int64             `json:"dateline"` // dayjs(dateline, 'X')
	TargetRow   *CoolapkTargetRow `json:"targetRow"`
}

// CoolapkResponse 酷安响应
type CoolapkResponse struct {
	Data []CoolapkDataItem `json:"data"`
}

// getRandomDEVICE_ID 获取随机设备ID
func getRandomDEVICE_ID() string {
	r := []int{10, 6, 6, 6, 14}
	id := make([]string, len(r))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	rand.Seed(time.Now().UnixNano())
	for i, length := range r {
		b := make([]byte, length)
		for j := range b {
			b[j] = charset[rand.Intn(len(charset))]
		}
		id[i] = string(b)
	}

	return strings.Join(id, "-")
}

// getAppToken 获取应用token（简化实现）
func getAppToken() string {
	// 注意：这里是简化实现，实际生产环境可能需要完整的加密逻辑
	DEVICE_ID := getRandomDEVICE_ID()
	now := time.Now().Unix()
	hexNow := fmt.Sprintf("0x%x", now)

	// 简化的token生成
	token := fmt.Sprintf("simplified_token_%s_%s", DEVICE_ID, hexNow)
	return token
}

// genHeaders 生成请求头
func genHeaders() http.Header {
	headers := http.Header{}
	headers.Set("X-Requested-With", "XMLHttpRequest")
	headers.Set("X-App-Id", "com.coolapk.market")
	headers.Set("X-App-Token", getAppToken())
	headers.Set("X-Sdk-Int", "29")
	headers.Set("X-Sdk-Locale", "zh-CN")
	headers.Set("X-App-Version", "11.0")
	headers.Set("X-Api-Version", "11")
	headers.Set("X-App-Code", "2101202")
	headers.Set("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 10; Redmi K30 5G MIUI/V12.0.3.0.QGICMXM) (#Build; Redmi; Redmi K30 5G; QKQ1.191222.002 test-keys; 10) +CoolMarket/11.0-2101202")

	return headers
}

// NewCoolapkSource 创建酷安数据源实例
func NewCoolapkSource() *CoolapkSource {
	return &CoolapkSource{
		BaseSource: BaseSource{
			Name:     "coolapk",
			URL:      "https://api.coolapk.com/v6/page/dataList?url=%2Ffeed%2FstatList%3FcacheExpires%3D300%26statType%3Dday%26sortField%3Ddetailnum%26title%3D%E4%BB%8A%E6%97%A5%E7%83%AD%E9%97%A8&title=%E4%BB%8A%E6%97%A5%E7%83%AD%E9%97%A8&subTitle=&page=1",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Fetch 获取酷安数据
func (s *CoolapkSource) Fetch(ctx context.Context) ([]byte, error) {
	// 创建HTTP请求
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	headers := genHeaders()
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// Parse 解析酷安热点内容
func (s *CoolapkSource) Parse(content []byte) ([]models.Item, error) {
	var resp CoolapkResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item

	for _, item := range resp.Data {
		if item.ID == "" {
			continue
		}

		// 获取标题
		title := item.EditorTitle
		if title == "" {
			// 从message中提取第一行作为标题
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Message))
			if err == nil {
				title = strings.Split(doc.Text(), "\n")[0]
			} else {
				// 简单处理，直接取文本第一行
				title = strings.Split(item.Message, "\n")[0]
			}
		}

		items = append(items, models.Item{
			ID:          item.ID,
			Title:       title,
			URL:         "https://www.coolapk.com" + item.URL,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewCoolapkSource())
}
