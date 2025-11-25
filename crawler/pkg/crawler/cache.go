package crawler

import (
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// Cache 定义了缓存的基本行为
type Cache interface {
	// Get 从缓存中获取数据
	Get(key string) ([]models.Item, error)

	// Set 将数据存入缓存
	Set(key string, items []models.Item, expiration time.Duration) error

	// Delete 从缓存中删除数据
	Delete(key string) error

	// Clear 清空缓存
	Clear() error
}
