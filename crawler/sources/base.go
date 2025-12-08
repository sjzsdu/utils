package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// BaseSource 是所有数据源的基础实现
type BaseSource struct {
	Name       string
	URL        string
	Interval   int
	Client     *http.Client
	Categories []string
}

// GetName 返回数据源名称
func (s *BaseSource) GetName() string {
	return s.Name
}

// GetURL 返回数据源的基础 URL
func (s *BaseSource) GetURL() string {
	return s.URL
}

// GetInterval 返回爬取间隔（秒）
func (s *BaseSource) GetInterval() int {
	return s.Interval
}

// Fetch 获取数据源内容
func (s *BaseSource) Fetch(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.GetURL(), nil)
	if err != nil {
		return nil, err
	}

	// 设置默认的User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Parse 解析获取到的内容，返回结构化数据
// 基础实现返回空数组，具体数据源需要重写此方法
func (s *BaseSource) Parse(content []byte) ([]models.Item, error) {
	return []models.Item{}, nil
}

// GetCategories 返回数据源的分类列表
func (s *BaseSource) GetCategories() []string {
	return s.Categories
}
