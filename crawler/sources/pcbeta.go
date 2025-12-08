package sources

import (
	"encoding/xml"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// PcbetaSource Pcbeta数据源
// 该数据源从两个RSS feed获取数据：
// 1. Windows 11 相关内容
// 2. Windows 相关内容
type PcbetaSource struct {
	BaseSource
}

// RSSItem 表示RSS中的条目
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// RSSFeed 表示整个RSS feed
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel RSSChannel `xml:"channel"`
}

// RSSChannel 表示RSS中的频道
type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

// NewPcbetaSource 创建Pcbeta数据源实例
func NewPcbetaSource() *PcbetaSource {
	return &PcbetaSource{
		BaseSource: BaseSource{
			Name:       "pcbeta",
			URL:        "https://bbs.pcbeta.com/forum.php?mod=rss&fid=563&auth=0",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技", "电脑"},
		},
	}
}

// Parse 解析Pcbeta RSS内容
func (s *PcbetaSource) Parse(content []byte) ([]models.Item, error) {
	var feed RSSFeed
	if err := xml.Unmarshal(content, &feed); err != nil {
		return nil, err
	}

	items := make([]models.Item, len(feed.Channel.Items))
	for i, rssItem := range feed.Channel.Items {
		// 解析发布时间
		pubTime, err := time.Parse(time.RFC1123, rssItem.PubDate)
		if err != nil {
			// 如果解析失败，使用当前时间
			pubTime = time.Now()
		}

		items[i] = models.Item{
			ID:          rssItem.GUID,
			Title:       rssItem.Title,
			URL:         rssItem.Link,
			Content:     rssItem.Description,
			Source:      s.Name,
			PublishedAt: pubTime,
			CreatedAt:   pubTime,
		}
	}

	return items, nil
}

func init() {
	RegisterSource(NewPcbetaSource())
}
