package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// Jin10Source 金十数据数据源
// 该数据源从金十数据网站获取财经快讯
type Jin10Source struct {
	BaseSource
}

// Jin10Item 金十数据项
type Jin10Item struct {
	ID   string `json:"id"`
	Time string `json:"time"`
	Type int    `json:"type"`
	Data struct {
		Pic        string `json:"pic,omitempty"`
		Title      string `json:"title,omitempty"`
		Source     string `json:"source,omitempty"`
		Content    string `json:"content,omitempty"`
		SourceLink string `json:"source_link,omitempty"`
		VipTitle   string `json:"vip_title,omitempty"`
		Lock       bool   `json:"lock"`
		VipLevel   int    `json:"vip_level"`
		VipDesc    string `json:"vip_desc,omitempty"`
	} `json:"data"`
	Important int      `json:"important"`
	Tags      []string `json:"tags"`
	Channel   []int    `json:"channel"`
	Remark    []any    `json:"remark"`
}

// NewJin10Source 创建金十数据数据源实例
func NewJin10Source() *Jin10Source {
	return &Jin10Source{
		BaseSource: BaseSource{
			Name:     "jin10",
			URL:      "https://www.jin10.com/flash_newest.js",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Fetch 获取金十数据财经快讯
func (s *Jin10Source) Fetch(ctx context.Context) ([]byte, error) {
	// 添加时间戳参数以避免缓存
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	originalURL := s.URL
	defer func() {
		// 恢复原始URL
		s.URL = originalURL
	}()

	// 临时修改URL以添加时间戳
	s.URL = fmt.Sprintf("%s?t=%d", originalURL, timestamp)
	return s.BaseSource.Fetch(ctx)
}

// Parse 解析金十数据财经快讯内容
func (s *Jin10Source) Parse(content []byte) ([]models.Item, error) {
	// 移除JavaScript变量声明，只保留JSON部分
	rawStr := string(content)
	jsonStr := strings.TrimSpace(strings.TrimRight(strings.TrimLeft(rawStr, "var newest ="), ";"))

	// 解析JSON数据
	var items []Jin10Item
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil, err
	}

	modelsItems := make([]models.Item, 0)

	// 过滤和转换数据
	for _, item := range items {
		// 过滤条件：有标题或内容，且不包含channel 5
		if item.Data.Title == "" && item.Data.Content == "" {
			continue
		}

		// 检查是否包含channel 5
		containsChannel5 := false
		for _, channel := range item.Channel {
			if channel == 5 {
				containsChannel5 = true
				break
			}
		}
		if containsChannel5 {
			continue
		}

		// 合并标题和内容
		text := item.Data.Title
		if text == "" {
			text = item.Data.Content
		}

		// 移除HTML标签
		text = strings.ReplaceAll(text, "<b>", "")
		text = strings.ReplaceAll(text, "</b>", "")

		// 提取标题和描述
		re := regexp.MustCompile(`^【([^】]*)】(.*)$`)
		matches := re.FindStringSubmatch(text)

		var title, desc string
		if len(matches) >= 3 {
			title = matches[1]
			desc = matches[2]
		} else {
			title = text
			desc = ""
		}

		// 解析发布时间
		pubTime, err := parseJin10Time(item.Time)
		if err != nil {
			pubTime = time.Now()
		}

		modelsItems = append(modelsItems, models.Item{
			ID:          item.ID,
			Title:       title,
			URL:         fmt.Sprintf("https://flash.jin10.com/detail/%s", item.ID),
			Content:     desc,
			Source:      s.Name,
			Category:    "财经",
			CreatedAt:   pubTime,
			PublishedAt: pubTime,
		})
	}

	return modelsItems, nil
}

// parseJin10Time 解析金十数据的时间格式
// 金十数据的时间格式类似于："08:30:22"
func parseJin10Time(timeStr string) (time.Time, error) {
	// 获取当前日期
	now := time.Now()
	year, month, day := now.Date()

	// 解析时间部分
	layout := "15:04:05"
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}

	// 合并日期和时间
	return time.Date(year, month, day, parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, now.Location()), nil
}

func init() {
	RegisterSource(NewJin10Source())
}
