package sources

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// MktNewsSource implements crawler.Source interface for MktNews

type MktNewsSource struct {
	BaseSource
}

// NewMktNewsSource creates a new MktNews source
func NewMktNewsSource() crawler.Source {
	return &MktNewsSource{
		BaseSource: BaseSource{
			Name:     "mktnews",
			URL:      "https://api.mktnews.net/api/flash/host",
			Interval: 300, // Interval in seconds
		},
	}
}

// MktNewsResponse represents the response structure from MktNews API
type MktNewsResponse struct {
	Data []CategoryData `json:"data"`
}

// CategoryData represents a category in the response
type CategoryData struct {
	Name  string            `json:"name"`
	Child []SubCategoryData `json:"child"`
}

// SubCategoryData represents a subcategory in the response
type SubCategoryData struct {
	FlashList []Report `json:"flash_list"`
}

// Report represents a news report in the response
type Report struct {
	ID   string     `json:"id"`
	Time string     `json:"time"`
	Type interface{} `json:"type"` // 使用interface{}类型以适应API返回的数字类型
	Data ReportData `json:"data"`
}

// ReportData represents the data of a news report
type ReportData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Pic     string `json:"pic"`
}

// Parse parses the JSON content from MktNews API and returns news items
func (s *MktNewsSource) Parse(content []byte) ([]models.Item, error) {
	var resp MktNewsResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Define categories to process
	categories := []string{"policy", "AI", "financial"}
	typeMap := map[string]string{
		"policy":    "Policy",
		"AI":        "AI",
		"financial": "Financial",
	}

	// Collect all reports from specified categories
	var allReports []Report
	for _, category := range categories {
		categoryData := findCategoryByName(resp.Data, category)
		if categoryData == nil || len(categoryData.Child) == 0 {
			continue
		}

		for _, subCategory := range categoryData.Child {
			for i := range subCategory.FlashList {
				subCategory.FlashList[i].Type = typeMap[category]
				allReports = append(allReports, subCategory.FlashList[i])
			}
		}
	}

	// Sort reports by time in descending order
	sort.Slice(allReports, func(i, j int) bool {
		return allReports[i].Time > allReports[j].Time
	})

	// Map reports to models.Item
	var items []models.Item
	for _, report := range allReports {
		// Generate title from content if title is empty
		title := report.Data.Title
		if title == "" {
			// Try to extract title from content using regex pattern
			if match := regexp.MustCompile(`^【([^】]*)】(.*)$`).FindStringSubmatch(report.Data.Content); match != nil {
				title = match[1]
			} else {
				title = report.Data.Content
			}
		}

		// Parse pubDate
		pubDate, err := time.Parse(time.RFC3339, report.Time)
		if err != nil {
			// If parsing fails, use current time as fallback
			pubDate = time.Now()
		}

		item := models.Item{
			ID:          report.ID,
			Title:       title,
			URL:         fmt.Sprintf("https://mktnews.net/flashDetail.html?id=%s", report.ID),
			Source:      s.Name,
			Content:     report.Data.Content,
			PublishedAt: pubDate,
		}
		items = append(items, item)
	}

	return items, nil
}

// findCategoryByName finds a category by name in the response data
func findCategoryByName(data []CategoryData, name string) *CategoryData {
	for i := range data {
		if data[i].Name == name {
			return &data[i]
		}
	}
	return nil
}

func init() {
	RegisterSource(NewMktNewsSource())
}
