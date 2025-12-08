package sources

import (
	"fmt"
	"sync"

	"github.com/sjzsdu/utils/crawler/pkg/crawler"
)

// Registry 是数据源注册表
type Registry struct {
	sources map[string]crawler.Source
	mu      sync.RWMutex
}

// registry 是全局数据源注册表实例
var (
	registry *Registry
	once     sync.Once
)

// GetRegistry 获取全局数据源注册表实例
func GetRegistry() *Registry {
	once.Do(func() {
		registry = &Registry{
			sources: make(map[string]crawler.Source),
		}
	})
	return registry
}

// Register 注册数据源
func (r *Registry) Register(source crawler.Source) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := source.GetName()
	if _, exists := r.sources[name]; exists {
		return fmt.Errorf("source %s already registered", name)
	}

	r.sources[name] = source
	return nil
}

// Get 获取数据源
func (r *Registry) Get(name string) (crawler.Source, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	source, exists := r.sources[name]
	if !exists {
		return nil, fmt.Errorf("source %s not found", name)
	}

	return source, nil
}

// List 列出所有数据源
func (r *Registry) List() []crawler.Source {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sources := make([]crawler.Source, 0, len(r.sources))
	for _, source := range r.sources {
		sources = append(sources, source)
	}

	return sources
}

// GetByCategory 根据类别获取数据源列表
func (r *Registry) GetByCategory(category string) []crawler.Source {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []crawler.Source
	for _, source := range r.sources {
		for _, c := range source.GetCategories() {
			if c == category {
				result = append(result, source)
				break
			}
		}
	}

	return result
}

// GetByCategories 根据多个类别获取数据源列表
func (r *Registry) GetByCategories(categories []string) []crawler.Source {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 创建类别映射，方便查找
	categoryMap := make(map[string]bool)
	for _, category := range categories {
		categoryMap[category] = true
	}

	// 用于去重的source名称映射
	seenSources := make(map[string]bool)
	var result []crawler.Source

	for _, source := range r.sources {
		for _, c := range source.GetCategories() {
			if categoryMap[c] && !seenSources[source.GetName()] {
				result = append(result, source)
				seenSources[source.GetName()] = true
				break
			}
		}
	}

	return result
}

// GetSources 根据名称列表获取多个数据源
func (r *Registry) GetSources(names []string) ([]crawler.Source, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []crawler.Source
	for _, name := range names {
		source, exists := r.sources[name]
		if !exists {
			return nil, fmt.Errorf("source %s not found", name)
		}
		result = append(result, source)
	}

	return result, nil
}

// RegisterSource 注册数据源到全局注册表
func RegisterSource(source crawler.Source) error {
	return GetRegistry().Register(source)
}
