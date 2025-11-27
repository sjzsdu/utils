package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// ClsSource 财联社数据源
type ClsSource struct {
	BaseSource
}

// ClsItem 财联社数据项
type ClsItem struct {
	ID       int64   `json:"id"`
	Title    *string `json:"title,omitempty"`
	Brief    string  `json:"brief"`
	Shareurl string  `json:"shareurl"`
	Ctime    int64   `json:"ctime"` // 需要 * 1000 转换为毫秒时间戳
	IsAd     int64   `json:"is_ad"` // 1 表示广告
}

// ClsTelegraphData 财联社电报数据
type ClsTelegraphData struct {
	RollData []ClsItem `json:"roll_data"`
}

// ClsTelegraphRes 财联社电报响应
type ClsTelegraphRes struct {
	Data ClsTelegraphData `json:"data"`
}

// ClsDepthData 财联社深度数据
type ClsDepthData struct {
	TopArticle []ClsItem `json:"top_article"`
	DepthList  []ClsItem `json:"depth_list"`
}

// ClsDepthRes 财联社深度响应
type ClsDepthRes struct {
	Data ClsDepthData `json:"data"`
}

// ClsHotRes 财联社热点响应
type ClsHotRes struct {
	Data []ClsItem `json:"data"`
}

// getClsSearchParams 获取财联社搜索参数
func getClsSearchParams() url.Values {
	params := url.Values{}
	params.Add("appName", "CailianpressWeb")
	params.Add("os", "web")
	params.Add("sv", "7.7.5")

	// 注意：原始代码中的sign生成需要加密，这里简化处理
	// 在实际生产环境中，可能需要实现完整的加密逻辑
	params.Add("sign", "simplified_sign")

	return params
}

// NewClsTelegraphSource 创建财联社电报数据源实例
func NewClsTelegraphSource() *ClsSource {
	return &ClsSource{
		BaseSource: BaseSource{
			Name:     "cls-telegraph",
			URL:      "https://www.cls.cn/nodeapi/updateTelegraphList",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// NewClsDepthSource 创建财联社深度数据源实例
func NewClsDepthSource() *ClsSource {
	return &ClsSource{
		BaseSource: BaseSource{
			Name:     "cls-depth",
			URL:      "https://www.cls.cn/v3/depth/home/assembled/1000",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// NewClsHotSource 创建财联社热点数据源实例
func NewClsHotSource() *ClsSource {
	return &ClsSource{
		BaseSource: BaseSource{
			Name:     "cls-hot",
			URL:      "https://www.cls.cn/v2/article/hot/list",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Fetch 获取财联社数据
func (s *ClsSource) Fetch(ctx context.Context) ([]byte, error) {
	// 创建URL并添加参数
	u, err := url.Parse(s.URL)
	if err != nil {
		return nil, err
	}

	// 添加查询参数
	u.RawQuery = getClsSearchParams().Encode()

	// 创建HTTP请求
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Parse 解析财联社内容
func (s *ClsSource) Parse(content []byte) ([]models.Item, error) {
	var items []models.Item

	switch s.Name {
	case "cls-telegraph":
		var resp ClsTelegraphRes
		if err := json.Unmarshal(content, &resp); err != nil {
			return nil, err
		}

		// 过滤非广告项
		for _, item := range resp.Data.RollData {
			if item.IsAd != 1 {
				title := item.Brief
				if item.Title != nil && *item.Title != "" {
					title = *item.Title
				}

				items = append(items, models.Item{
					ID:          fmt.Sprintf("%d", item.ID),
					Title:       title,
					URL:         fmt.Sprintf("https://www.cls.cn/detail/%d", item.ID),
					Source:      s.Name,
					CreatedAt:   time.Now(),
					PublishedAt: time.Unix(item.Ctime, 0).UTC(),
				})
			}
		}

	case "cls-depth":
		var resp ClsDepthRes
		if err := json.Unmarshal(content, &resp); err != nil {
			return nil, err
		}

		// 按时间排序
		sort.Slice(resp.Data.DepthList, func(i, j int) bool {
			return resp.Data.DepthList[i].Ctime > resp.Data.DepthList[j].Ctime
		})

		for _, item := range resp.Data.DepthList {
			title := item.Brief
			if item.Title != nil && *item.Title != "" {
				title = *item.Title
			}

			items = append(items, models.Item{
				ID:          fmt.Sprintf("%d", item.ID),
				Title:       title,
				URL:         fmt.Sprintf("https://www.cls.cn/detail/%d", item.ID),
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Unix(item.Ctime, 0).UTC(),
			})
		}

	case "cls-hot":
		var resp ClsHotRes
		if err := json.Unmarshal(content, &resp); err != nil {
			return nil, err
		}

		for _, item := range resp.Data {
			title := item.Brief
			if item.Title != nil && *item.Title != "" {
				title = *item.Title
			}

			items = append(items, models.Item{
				ID:          fmt.Sprintf("%d", item.ID),
				Title:       title,
				URL:         fmt.Sprintf("https://www.cls.cn/detail/%d", item.ID),
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Now(),
			})
		}
	}

	return items, nil
}

func init() {
	// 注册三个子数据源
	RegisterSource(NewClsTelegraphSource())
	RegisterSource(NewClsDepthSource())
	RegisterSource(NewClsHotSource())

	// 注册默认的cls数据源，指向telegraph
	clsSource := NewClsTelegraphSource()
	clsSource.Name = "cls"
	RegisterSource(clsSource)
}
