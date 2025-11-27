package sources

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// DoubanSource 豆瓣数据源
type DoubanSource struct {
	BaseSource
}

// DoubanMovieItem 豆瓣电影项
type DoubanMovieItem struct {
	Rating struct {
		Count     int     `json:"count"`
		Max       int     `json:"max"`
		StarCount int     `json:"star_count"`
		Value     float64 `json:"value"`
	} `json:"rating"`
	Title string `json:"title"`
	Pic   struct {
		Large  string `json:"large"`
		Normal string `json:"normal"`
	} `json:"pic"`
	IsNew        bool   `json:"is_new"`
	URI          string `json:"uri"`
	EpisodesInfo string `json:"episodes_info"`
	CardSubtitle string `json:"card_subtitle"`
	Type         string `json:"type"`
	ID           string `json:"id"`
}

// DoubanResult 豆瓣API响应结构
type DoubanResult struct {
	Category      string            `json:"category"`
	Tags          []interface{}     `json:"tags"`
	Items         []DoubanMovieItem `json:"items"`
	RecommendTags []interface{}     `json:"recommend_tags"`
	Total         int               `json:"total"`
	Type          string            `json:"type"`
}

// NewDoubanSource 创建豆瓣数据源实例
func NewDoubanSource() *DoubanSource {
	return &DoubanSource{
		BaseSource: BaseSource{
			Name:     "douban",
			URL:      "https://m.douban.com/rexxar/api/v2/subject/recent_hot/movie",
			Interval: 3600, // 1小时爬取一次
		},
	}
}

// Parse 解析豆瓣热门电影内容
func (s *DoubanSource) Parse(content []byte) ([]models.Item, error) {
	var result DoubanResult
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, err
	}

	items := make([]models.Item, 0, len(result.Items))
	now := time.Now()
	for _, item := range result.Items {
		items = append(items, models.Item{
			ID:          item.ID,
			Title:       item.Title,
			URL:         fmt.Sprintf("https://movie.douban.com/subject/%s", item.ID),
			Content:     item.CardSubtitle,
			Source:      s.Name,
			Category:    "movie",
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewDoubanSource())
}
