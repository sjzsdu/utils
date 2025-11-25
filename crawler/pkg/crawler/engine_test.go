package crawler_test

import (
	"context"
	"testing"
	"time"

	"github.com/sjzsdu/utils/crawler/internal/cache"
	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// mockSource 是一个用于测试的模拟数据源
type mockSource struct {
	name     string
	url      string
	interval int
	items    []models.Item
}

func (m *mockSource) GetName() string {
	return m.name
}

func (m *mockSource) GetURL() string {
	return m.url
}

func (m *mockSource) Fetch(ctx context.Context) ([]byte, error) {
	return []byte("test content"), nil
}

func (m *mockSource) Parse(content []byte) ([]models.Item, error) {
	return m.items, nil
}

func (m *mockSource) GetInterval() int {
	return m.interval
}

func TestEngine(t *testing.T) {
	// 创建内存缓存
	memCache := cache.NewMemoryCache(1 * time.Hour)
	defer memCache.Close()

	// 创建爬取引擎
	engine := crawler.NewEngine(memCache)

	// 创建模拟数据源
	mockItems := []models.Item{
		{
			ID:    "test-1",
			Title: "Test Item 1",
			URL:   "https://example.com/1",
		},
	}

	source := &mockSource{
		name:     "test",
		url:      "https://example.com",
		interval: 60,
		items:    mockItems,
	}

	// 测试注册数据源
	if err := engine.RegisterSource(source); err != nil {
		t.Errorf("Failed to register source: %v", err)
	}

	// 测试获取数据源
	items, err := engine.FetchItem(context.Background(), "test")
	if err != nil {
		t.Errorf("Failed to fetch item: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}

	if items[0].Title != "Test Item 1" {
		t.Errorf("Expected title 'Test Item 1', got '%s'", items[0].Title)
	}

	// 测试注销数据源
	if err := engine.UnregisterSource("test"); err != nil {
		t.Errorf("Failed to unregister source: %v", err)
	}

	// 测试订阅功能
	if err := engine.RegisterSource(source); err != nil {
		t.Errorf("Failed to register source again: %v", err)
	}

	ch := make(chan []models.Item, 10)
	if err := engine.Subscribe("test", ch); err != nil {
		t.Errorf("Failed to subscribe: %v", err)
	}

	if err := engine.Unsubscribe("test", ch); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}

	// 测试启动和停止引擎
	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Errorf("Failed to start engine: %v", err)
	}

	if err := engine.Stop(); err != nil {
		t.Errorf("Failed to stop engine: %v", err)
	}
}
