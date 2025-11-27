package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sjzsdu/utils/crawler/internal/cache"
	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/pkg/models"
	"github.com/sjzsdu/utils/crawler/sources"
)

func main() {
	// 创建内存缓存
	memCache := cache.NewMemoryCache(1 * time.Hour)
	defer memCache.Close()

	// 创建爬取引擎
	engine := crawler.NewEngine(memCache)

	// 获取数据源注册表
	registry := sources.GetRegistry()

	// 注册所有数据源
	for _, source := range registry.List() {
		if err := engine.RegisterSource(source); err != nil {
			fmt.Printf("Failed to register source %s: %v\n", source.GetName(), err)
			return
		}
		fmt.Printf("Registered source: %s\n", source.GetName())
	}

	// 启动爬取引擎
	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		fmt.Printf("Failed to start engine: %v\n", err)
		return
	}
	defer engine.Stop()

	// 订阅 GitHub 数据源的更新
	githubChan := make(chan []models.Item, 10)
	if err := engine.Subscribe("github", githubChan); err != nil {
		fmt.Printf("Failed to subscribe to github: %v\n", err)
		return
	}
	defer engine.Unsubscribe("github", githubChan)

	fmt.Println("Crawler engine started. Waiting for updates...")
	fmt.Println("Press Ctrl+C to exit.")

	// 处理爬取结果
	for {
		select {
		case items := <-githubChan:
			fmt.Printf("\nReceived %d items from GitHub\n", len(items))
			// 只显示前5条或实际数量，避免切片越界
			maxItems := 5
			if len(items) < maxItems {
				maxItems = len(items)
			}
			for i, item := range items[:maxItems] {
				fmt.Printf("%d. %s\n", i+1, item.Title)
				fmt.Printf("   URL: %s\n", item.URL)
				fmt.Printf("   Category: %s\n", item.Category)
			}
		case <-time.After(30 * time.Minute):
			fmt.Println("Timeout, exiting...")
			return
		}
	}
}
