package sources

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// GHXISource implements crawler.Source interface for GHXI news
type GHXISource struct {
	BaseSource
}

// NewGHXISource creates a new GHXI source
func NewGHXISource() crawler.Source {
	return &GHXISource{
		BaseSource: BaseSource{
			Name:     "ghxi",
			URL:      "https://www.ghxi.com/category/all",
			Interval: 300, // Interval in seconds
		},
	}
}

// Parse parses the HTML content from GHXI website and returns news items
func (s *GHXISource) Parse(content []byte) ([]models.Item, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var items []models.Item
	doc.Find(".sec-panel .sec-panel-body .post-loop li").Each(func(i int, sel *goquery.Selection) {
		title := sel.Find(".item-content .item-title").Text()
		title = strings.TrimSpace(title)
		title = strings.ReplaceAll(title, "'", "''")

		description := sel.Find(".item-content .item-excerpt").Text()
		description = strings.TrimSpace(description)
		description = strings.ReplaceAll(description, "'", "''")

		dateStr := sel.Find(".item-content .date").Text()
		pubDate := s.parseRelativeTime(dateStr)

		url, exists := sel.Find(".item-content .item-title a").Attr("href")
		if exists && url != "" {
			item := models.Item{
				ID:          url,
				Title:       title,
				URL:         url,
				Source:      s.Name,
				Content:     description,
				PublishedAt: time.Unix(pubDate, 0),
			}
			items = append(items, item)
		}
	})

	return items, nil
}

// parseRelativeTime parses relative time strings like "3 hours ago" to timestamp (Unix seconds)
func (s *GHXISource) parseRelativeTime(timeStr string) int64 {
	re := regexp.MustCompile(`^(\d+)\s*([秒天周月年]|分钟|小时)`)
	matches := re.FindStringSubmatch(timeStr)
	if matches == nil || len(matches) != 3 {
		return 0
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "秒":
		duration = time.Duration(num) * time.Second
	case "分钟":
		duration = time.Duration(num) * time.Minute
	case "小时":
		duration = time.Duration(num) * time.Hour
	case "天":
		duration = time.Duration(num) * 24 * time.Hour
	case "周":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "月":
		duration = time.Duration(num) * 30 * 24 * time.Hour // Approximate
	case "年":
		duration = time.Duration(num) * 365 * 24 * time.Hour // Approximate
	default:
		return 0
	}

	return time.Now().Add(-duration).Unix()
}

func init() {
	RegisterSource(NewGHXISource())
}
