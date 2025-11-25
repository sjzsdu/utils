package cache

import (
	"sync"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// item 表示缓存中的一个条目
type item struct {
	value      []models.Item
	expiration int64
}

// MemoryCache 是基于内存的缓存实现
type MemoryCache struct {
	items    map[string]item
	mu       sync.RWMutex
	gcTicker *time.Ticker
	stopChan chan struct{}
}

// NewMemoryCache 创建一个新的内存缓存实例
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items:    make(map[string]item),
		gcTicker: time.NewTicker(cleanupInterval),
		stopChan: make(chan struct{}),
	}

	// 启动垃圾回收协程
	go cache.gc()

	return cache
}

// gc 定期清理过期的缓存项
func (c *MemoryCache) gc() {
	for {
		select {
		case <-c.gcTicker.C:
			c.deleteExpired()
		case <-c.stopChan:
			c.gcTicker.Stop()
			return
		}
	}
}

// deleteExpired 删除所有过期的缓存项
func (c *MemoryCache) deleteExpired() {
	now := time.Now().UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if v.expiration > 0 && now > v.expiration {
			delete(c.items, k)
		}
	}
}

// Get 从缓存中获取数据
func (c *MemoryCache) Get(key string) ([]models.Item, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, nil
	}

	// 检查是否过期
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil, nil
	}

	return item.value, nil
}

// Set 将数据存入缓存
func (c *MemoryCache) Set(key string, value []models.Item, expiration time.Duration) error {
	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:      value,
		expiration: exp,
	}

	return nil
}

// Delete 从缓存中删除数据
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Clear 清空缓存
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]item)
	return nil
}

// Close 关闭缓存，停止垃圾回收
func (c *MemoryCache) Close() {
	close(c.stopChan)
}
