package sources

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// LinuxdoTopic LinuxDo话题
type LinuxdoTopic struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	FancyTitle         string    `json:"fancy_title"`
	PostsCount         int       `json:"posts_count"`
	ReplyCount         int       `json:"reply_count"`
	HighestPostNumber  int       `json:"highest_post_number"`
	ImageURL           *string   `json:"image_url"`
	CreatedAt          time.Time `json:"created_at"`
	LastPostedAt       time.Time `json:"last_posted_at"`
	Bumped             bool      `json:"bumped"`
	BumpedAt           time.Time `json:"bumped_at"`
	Unseen             bool      `json:"unseen"`
	Pinned             bool      `json:"pinned"`
	Excerpt            *string   `json:"excerpt"`
	Visible            bool      `json:"visible"`
	Closed             bool      `json:"closed"`
	Archived           bool      `json:"archived"`
	LikeCount          int       `json:"like_count"`
	HasSummary         bool      `json:"has_summary"`
	LastPosterUsername string    `json:"last_poster_username"`
	CategoryID         int       `json:"category_id"`
	PinnedGlobally     bool      `json:"pinned_globally"`
}

// LinuxdoResponse LinuxDo响应
type LinuxdoResponse struct {
	TopicList struct {
		CanCreateTopic   bool            `json:"can_create_topic"`
		MoreTopicsURL    string          `json:"more_topics_url"`
		PerPage          int             `json:"per_page"`
		TopTags          []string        `json:"top_tags"`
		Topics           []LinuxdoTopic  `json:"topics"`
	} `json:"topic_list"`
}

// LinuxdoHotSource LinuxDo热门话题数据源
type LinuxdoHotSource struct {
	BaseSource
}

// LinuxdoLatestSource LinuxDo最新话题数据源
type LinuxdoLatestSource struct {
	BaseSource
}

// NewLinuxdoHotSource 创建LinuxDo热门话题数据源实例
func NewLinuxdoHotSource() *LinuxdoHotSource {
	return &LinuxdoHotSource{
		BaseSource: BaseSource{
			Name:       "linuxdo-hot",
			URL:        "https://linux.do/top/daily.json",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技"},
		},
	}
}

// NewLinuxdoLatestSource 创建LinuxDo最新话题数据源实例
func NewLinuxdoLatestSource() *LinuxdoLatestSource {
	return &LinuxdoLatestSource{
		BaseSource: BaseSource{
			Name:       "linuxdo-latest",
			URL:        "https://linux.do/latest.json?order=created",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"科技"},
		},
	}
}

// Fetch 获取LinuxDo热门话题数据
func (s *LinuxdoHotSource) Fetch(ctx context.Context) ([]byte, error) {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头，模拟浏览器访问
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	client := s.Client
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Parse 解析LinuxDo热门话题内容
func (s *LinuxdoHotSource) Parse(content []byte) ([]models.Item, error) {
	// 检查是否为HTML响应
	if len(content) > 0 && content[0] == '<' {
		// 如果是HTML，返回空数组，避免JSON解析错误
		return []models.Item{}, nil
	}
	
	var resp LinuxdoResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		// JSON解析失败，返回空数组而非错误
		return []models.Item{}, nil
	}

	var items []models.Item
	// 过滤并转换话题
	for _, topic := range resp.TopicList.Topics {
		// 过滤掉不可见、归档和置顶的内容
		if !topic.Visible || topic.Archived || topic.Pinned {
			continue
		}

		items = append(items, models.Item{
			ID:          strconv.Itoa(topic.ID),
			Title:       topic.Title,
			URL:         "https://linux.do/t/topic/" + strconv.Itoa(topic.ID),
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

// Parse 解析LinuxDo最新话题内容
func (s *LinuxdoLatestSource) Parse(content []byte) ([]models.Item, error) {
	var resp LinuxdoResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var items []models.Item
	// 过滤并转换话题
	for _, topic := range resp.TopicList.Topics {
		// 过滤掉不可见、归档和置顶的内容
		if !topic.Visible || topic.Archived || topic.Pinned {
			continue
		}

		items = append(items, models.Item{
			ID:          strconv.Itoa(topic.ID),
			Title:       topic.Title,
			URL:         "https://linux.do/t/topic/" + strconv.Itoa(topic.ID),
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: topic.CreatedAt,
		})
	}

	return items, nil
}

// LinuxdoSource LinuxDo主数据源
type LinuxdoSource struct {
	LinuxdoLatestSource
}

// NewLinuxdoSource 创建LinuxDo主数据源实例
func NewLinuxdoSource() *LinuxdoSource {
	source := &LinuxdoSource{
		LinuxdoLatestSource: *NewLinuxdoLatestSource(),
	}
	source.Name = "linuxdo"
	return source
}

func init() {
	// 注册三个数据源：主数据源、最新数据源、热门数据源
	RegisterSource(NewLinuxdoSource())
	RegisterSource(NewLinuxdoLatestSource())
	RegisterSource(NewLinuxdoHotSource())
}
