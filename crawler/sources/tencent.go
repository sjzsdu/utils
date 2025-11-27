package sources

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// TencentSource 腾讯新闻数据源
type TencentSource struct {
	BaseSource
}

// TencentTabItem 腾讯新闻标签项
type TencentTabItem struct {
	ID           string        `json:"id"`
	ChannelID    string        `json:"channel_id"`
	Name         string        `json:"name"`
	Source       string        `json:"source"`
	Type         string        `json:"type"`
	ArticleList  []interface{} `json:"articleList"`
	ArticleCount int           `json:"article_count"`
	SubTab       string        `json:"sub_tab"`
}

// TencentHeadArticle 腾讯新闻头条文章
type TencentHeadArticle struct {
	LiveInfo  string `json:"live_info"`
	Title     string `json:"title"`
	Img       string `json:"img"`
	PubTime   string `json:"pub_time"`
	MediaName string `json:"media_name"`
}

// TencentData 腾讯新闻数据
type TencentData struct {
	ID            int                `json:"id"`
	Name          string             `json:"name"`
	Lead          string             `json:"lead"`
	Cover         *string            `json:"cover"`
	ShareTitle    string             `json:"shareTitle"`
	ShareAbstract string             `json:"shareAbstract"`
	SharePic      string             `json:"sharePic"`
	Is724         bool               `json:"is724"`
	Is724Paper    bool               `json:"is724Paper"`
	HeadCMSID     string             `json:"head_cmsid"`
	FeedStyle     int                `json:"feed_style"`
	HeadArticle   TencentHeadArticle `json:"head_article"`
	PaperInfo     interface{}        `json:"paperInfo"`
	Tabs          []TencentTabItem   `json:"tabs"`
	Banner        string             `json:"banner"`
}

// TencentResponse 腾讯新闻响应
type TencentResponse struct {
	Ret  int         `json:"ret"`
	Msg  string      `json:"msg"`
	Data TencentData `json:"data"`
}

// NewTencentSource 创建腾讯新闻数据源实例
func NewTencentSource() *TencentSource {
	return &TencentSource{
		BaseSource: BaseSource{
			Name:     "tencent-hot",
			URL:      "https://i.news.qq.com/web_backend/v2/getTagInfo?tagId=aEWqxLtdgmQ%3D",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Fetch 获取腾讯新闻数据
func (s *TencentSource) Fetch(ctx context.Context) ([]byte, error) {
	// 创建HTTP请求
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置Referer头
	req.Header.Set("Referer", "https://news.qq.com/")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Parse 解析腾讯新闻内容
func (s *TencentSource) Parse(content []byte) ([]models.Item, error) {
	var resp TencentResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item

	// 检查是否有tabs
	if len(resp.Data.Tabs) > 0 {
		// 获取第一个tab的文章列表
		firstTab := resp.Data.Tabs[0]

		// 遍历文章列表
		for _, articleInterface := range firstTab.ArticleList {
			article, ok := articleInterface.(map[string]interface{})
			if !ok {
				continue
			}

			// 提取文章信息
			id, _ := article["id"].(string)
			title, _ := article["title"].(string)

			// 提取链接信息
			linkInfo, ok := article["link_info"].(map[string]interface{})
			if !ok {
				continue
			}
			url, _ := linkInfo["url"].(string)

			if id != "" && title != "" && url != "" {
				items = append(items, models.Item{
					ID:          id,
					Title:       title,
					URL:         url,
					Source:      s.Name,
					CreatedAt:   time.Now(),
					PublishedAt: time.Now(),
				})
			}
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewTencentSource())
}
