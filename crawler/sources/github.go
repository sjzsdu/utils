package sources

import (
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// GitHubSource GitHub数据源
type GitHubSource struct {
	BaseSource
}

// NewGitHubSource 创建GitHub数据源实例
func NewGitHubSource() *GitHubSource {
	return &GitHubSource{
		BaseSource: BaseSource{
			Name:       "github",
			URL:        "https://github.com/trending",
			Interval:   3600, // 1小时爬取一次
			Categories: []string{"科技", "编程"},
		},
	}
}

// Parse 解析GitHub趋势内容
func (s *GitHubSource) Parse(content []byte) ([]models.Item, error) {
	// 简单实现，返回空列表
	// 后续可以添加完整的HTML解析逻辑
	return []models.Item{}, nil
}

func init() {
	RegisterSource(NewGitHubSource())
}
