package sources

import (
	"bytes"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// SteamSource Steam数据源
// 该数据源从Steam获取热门游戏和玩家数量信息
type SteamSource struct {
	BaseSource
}

// NewSteamSource 创建Steam数据源实例
func NewSteamSource() *SteamSource {
	return &SteamSource{
		BaseSource: BaseSource{
			Name:       "steam",
			URL:        "https://store.steampowered.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"游戏", "科技"},
		},
	}
}

// Parse 解析Steam热门游戏内容
func (s *SteamSource) Parse(content []byte) ([]models.Item, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	var items []models.Item

	doc.Find("#detailStats tr.player_count_row").Each(func(i int, selection *goquery.Selection) {
		link := selection.Find("a.gameLink")
		url, _ := link.Attr("href")
		gameName := link.Text()
		currentPlayers := selection.Find("td:first-child .currentServers").Text()

		if url != "" && gameName != "" && currentPlayers != "" {
			items = append(items, models.Item{
				ID:          url,
				Title:       gameName,
				URL:         url,
				Source:      s.Name,
				CreatedAt:   time.Now(),
				PublishedAt: time.Now(),
			})
		}
	})

	return items, nil
}

func init() {
	RegisterSource(NewSteamSource())
}
