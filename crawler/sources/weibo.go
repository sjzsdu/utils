package sources

import (
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// WeiboSource 微博数据源
type WeiboSource struct {
	BaseSource
}

// NewWeiboSource 创建微博数据源实例
func NewWeiboSource() *WeiboSource {
	return &WeiboSource{
		BaseSource: BaseSource{
			Name:     "weibo",
			URL:      "https://s.weibo.com/top/summary?cate=realtimehot",
			Interval: 300, // 5分钟爬取一次
		},
	}
}

// Parse 解析微博热搜内容
func (s *WeiboSource) Parse(content []byte) ([]models.Item, error) {
	// 简单实现，返回空列表
	// 后续可以添加完整的HTML解析逻辑
	return []models.Item{}, nil
}

func init() {
	RegisterSource(NewWeiboSource())
}
