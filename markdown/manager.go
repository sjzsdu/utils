package markdown

import (
	"sync"
)

// Manager 定义Markdown内容管理的接口
type Manager interface {
	// AddContent 添加Markdown内容
	AddContent(path, content string) error
	// GetContent 获取Markdown内容
	GetContent(path string) (string, bool)
	// UpdateContent 更新Markdown内容
	UpdateContent(path, content string) error
	// DeleteContent 删除Markdown内容
	DeleteContent(path string) error
	// ListContents 列出所有Markdown内容的路径
	ListContents() []string
	// GetAllContent 获取所有Markdown内容
	GetAllContent() map[string]string
	// Clear 清空所有Markdown内容
	Clear()
}

// MarkdownManager 实现Manager接口，管理Markdown内容的map
type MarkdownManager struct {
	contentMap map[string]string // key: 文件路径, value: Markdown内容
	mu         sync.RWMutex
}

// NewMarkdownManager 创建新的Markdown内容管理器
func NewMarkdownManager() *MarkdownManager {
	return &MarkdownManager{
		contentMap: make(map[string]string),
	}
}

// AddContent 添加Markdown内容
func (m *MarkdownManager) AddContent(path, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contentMap[path] = content
	return nil
}

// GetContent 获取Markdown内容
func (m *MarkdownManager) GetContent(path string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	content, exists := m.contentMap[path]
	return content, exists
}

// UpdateContent 更新Markdown内容
func (m *MarkdownManager) UpdateContent(path, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.contentMap[path]; !exists {
		return ErrContentNotFound
	}
	m.contentMap[path] = content
	return nil
}

// DeleteContent 删除Markdown内容
func (m *MarkdownManager) DeleteContent(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.contentMap[path]; !exists {
		return ErrContentNotFound
	}
	delete(m.contentMap, path)
	return nil
}

// ListContents 列出所有Markdown内容的路径
func (m *MarkdownManager) ListContents() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	paths := make([]string, 0, len(m.contentMap))
	for path := range m.contentMap {
		paths = append(paths, path)
	}
	return paths
}

// GetAllContent 获取所有Markdown内容
func (m *MarkdownManager) GetAllContent() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// 创建一个新的map副本，避免外部修改内部状态
	content := make(map[string]string, len(m.contentMap))
	for path, contentStr := range m.contentMap {
		content[path] = contentStr
	}
	return content
}

// Clear 清空所有Markdown内容
func (m *MarkdownManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contentMap = make(map[string]string)
}
