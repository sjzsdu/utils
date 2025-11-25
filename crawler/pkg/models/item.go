package models

import (
	"time"
)

// Item 定义了爬取到的数据项
type Item struct {
	// ID 唯一标识符
	ID string `json:"id"`

	// Title 标题
	Title string `json:"title"`

	// URL 链接
	URL string `json:"url"`

	// Content 内容
	Content string `json:"content"`

	// Source 数据源名称
	Source string `json:"source"`

	// Category 分类
	Category string `json:"category"`

	// Images 图片链接列表
	Images []string `json:"images"`

	// PublishedAt 发布时间
	PublishedAt time.Time `json:"published_at"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// Result 定义了爬取结果
type Result struct {
	// Source 数据源名称
	Source string `json:"source"`

	// Items 爬取到的数据项列表
	Items []Item `json:"items"`

	// Success 是否成功
	Success bool `json:"success"`

	// Error 错误信息
	Error string `json:"error,omitempty"`

	// Timestamp 爬取时间
	Timestamp time.Time `json:"timestamp"`
}
