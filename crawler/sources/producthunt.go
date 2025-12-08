package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// ProducthuntSource Product Hunt数据源
// 该数据源从Product Hunt获取热门产品
// 注意：需要设置环境变量PRODUCTHUNT_API_TOKEN

type ProducthuntSource struct {
	BaseSource
}

// ProducthuntPost Product Hunt产品项
type ProducthuntPost struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Tagline    string `json:"tagline"`
	VotesCount int    `json:"votesCount"`
	URL        string `json:"url"`
	Slug       string `json:"slug"`
}

// ProducthuntEdge Product Hunt边
type ProducthuntEdge struct {
	Node ProducthuntPost `json:"node"`
}

// ProducthuntPosts Product Hunt帖子
type ProducthuntPosts struct {
	Edges []ProducthuntEdge `json:"edges"`
}

// ProducthuntData Product Hunt数据
type ProducthuntData struct {
	Posts ProducthuntPosts `json:"posts"`
}

// ProducthuntResponse Product Hunt响应
type ProducthuntResponse struct {
	Data ProducthuntData `json:"data"`
}

// NewProducthuntSource 创建Product Hunt数据源实例
func NewProducthuntSource() *ProducthuntSource {
	return &ProducthuntSource{
		BaseSource: BaseSource{
			Name:       "producthunt",
			URL:        "https://api.producthunt.com/v2/api/graphql",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技", "产品"},
		},
	}
}

// Fetch 获取Product Hunt热门产品数据
func (s *ProducthuntSource) Fetch(ctx context.Context) ([]byte, error) {
	apiToken := os.Getenv("PRODUCTHUNT_API_TOKEN")
	if apiToken == "" {
		return nil, nil // 如果没有API令牌，返回空结果
	}

	query := `
    query {
      posts(first: 30, order: VOTES) {
        edges {
          node {
            id
            name
            tagline
            votesCount
            url
            slug
          }
        }
      }
    }
  `

	// 构建请求体
	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return nil, err
	}

	// 创建HTTP请求
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// Parse 解析Product Hunt热门产品内容
func (s *ProducthuntSource) Parse(content []byte) ([]models.Item, error) {
	if len(content) == 0 {
		return []models.Item{}, nil // 如果内容为空，返回空结果
	}

	var resp ProducthuntResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item
	for _, edge := range resp.Data.Posts.Edges {
		post := edge.Node
		if post.ID != "" && post.Name != "" {
			url := post.URL
			if url == "" {
				url = "https://www.producthunt.com/posts/" + post.Slug
			}

			items = append(items, models.Item{
				ID:          post.ID,
				Title:       post.Name,
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
	RegisterSource(NewProducthuntSource())
}
