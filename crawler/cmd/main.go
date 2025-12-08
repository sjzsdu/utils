package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sjzsdu/utils/crawler/internal/cache"
	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/pkg/models"
	"github.com/sjzsdu/utils/crawler/sources"
)

func main() {
	// 解析命令行参数
	var categoriesStr, sourcesStr string
	flag.StringVar(&categoriesStr, "categories", "", "指定要爬取的类别列表，多个类别用逗号分隔")
	flag.StringVar(&sourcesStr, "sources", "", "指定要爬取的数据源名称列表，多个名称用逗号分隔")
	flag.Parse()

	// 创建日志文件
	logFile, err := os.OpenFile("crawler_results.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	// 创建内存缓存
	memCache := cache.NewMemoryCache(1 * time.Hour)
	defer memCache.Close()

	// 创建爬取引擎
	engine := crawler.NewEngine(memCache)

	// 获取数据源注册表
	registry := sources.GetRegistry()
	var selectedSources []crawler.Source

	// 根据命令行参数获取要爬取的数据源
	if sourcesStr != "" {
		// 根据数据源名称获取
		sourceNames := strings.Split(sourcesStr, ",")
		for i := range sourceNames {
			sourceNames[i] = strings.TrimSpace(sourceNames[i])
		}
		var err error
		selectedSources, err = registry.GetSources(sourceNames)
		if err != nil {
			fmt.Printf("Failed to get sources: %v\n", err)
			return
		}
	} else if categoriesStr != "" {
		// 根据类别获取
		categories := strings.Split(categoriesStr, ",")
		for i := range categories {
			categories[i] = strings.TrimSpace(categories[i])
		}
		selectedSources = registry.GetByCategories(categories)
		if len(selectedSources) == 0 {
			fmt.Printf("No sources found for categories: %s\n", categoriesStr)
			return
		}
	} else {
		// 爬取所有数据源
		selectedSources = registry.List()
	}

	// 注册所有数据源
	for _, source := range selectedSources {
		if err := engine.RegisterSource(source); err != nil {
			fmt.Printf("Failed to register source %s: %v\n", source.GetName(), err)
			return
		}
		fmt.Printf("Registered source: %s\n", source.GetName())
		logger.Printf("Registered source: %s\n", source.GetName())
	}

	// 启动爬取引擎
	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		fmt.Printf("Failed to start engine: %v\n", err)
		logger.Printf("Failed to start engine: %v\n", err)
		return
	}
	defer engine.Stop()

	// 为每个数据源创建订阅通道
	sourceChannels := make(map[string]chan []models.Item)
	for _, source := range selectedSources {
		sourceName := source.GetName()
		sourceChannels[sourceName] = make(chan []models.Item, 10)
		if err := engine.Subscribe(sourceName, sourceChannels[sourceName]); err != nil {
			fmt.Printf("Failed to subscribe to %s: %v\n", sourceName, err)
			logger.Printf("Failed to subscribe to %s: %v\n", sourceName, err)
			continue
		}
		defer engine.Unsubscribe(sourceName, sourceChannels[sourceName])
	}

	fmt.Println("Crawler engine started. Waiting for updates...")
	fmt.Println("Press Ctrl+C to exit.")
	logger.Println("Crawler engine started. Waiting for updates...")

	// 用于跟踪已处理的数据源数量
	processedSources := make(map[string]bool)
	totalSources := len(selectedSources)

	// 用于统计每个数据源获取的数据量
	resultsStats := make(map[string]int)

	// 创建一个统一的结果通道
	resultChan := make(chan struct {
		sourceName string
		items      []models.Item
	}, 100)

	// 为每个数据源启动一个goroutine处理结果
	for _, source := range selectedSources {
		sourceName := source.GetName()
		go func(name string, ch chan []models.Item) {
			// 设置10秒超时，避免无限期等待
			select {
			case items := <-ch:
				resultChan <- struct {
					sourceName string
					items      []models.Item
				}{name, items}
			case <-time.After(10 * time.Second):
				// 超时处理，返回空结果
				resultChan <- struct {
					sourceName string
					items      []models.Item
				}{name, []models.Item{}}
			}
		}(sourceName, sourceChannels[sourceName])
	}

	// 处理爬取结果的主循环
	timeout := time.After(10 * time.Minute)
	checkInterval := time.Tick(5 * time.Second)

	for {
		select {
		case result := <-resultChan:
			// 记录结果
			resultsStats[result.sourceName] = len(result.items)
			logSourceResults(result.sourceName, result.items, logger)
			processedSources[result.sourceName] = true

		// 检查是否所有数据源都已处理
		case <-checkInterval:
			if len(processedSources) >= totalSources {
				fmt.Println("All sources processed, exiting...")
				logger.Println("All sources processed, exiting...")
				goto ExitLoop
			}

		// 超时退出
		case <-timeout:
			fmt.Println("Timeout, exiting...")
			logger.Printf("Timeout, exiting. Processed %d/%d sources\n", len(processedSources), totalSources)
			goto ExitLoop
		}
	}

ExitLoop:
	// 打印统计结果
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("Crawler Results Summary")
	fmt.Println(separator)
	logger.Println("\n" + separator)
	logger.Println("Crawler Results Summary")
	logger.Println(separator)

	// 按照数据源名称排序
	sourceNames := make([]string, 0, len(selectedSources))
	for _, source := range selectedSources {
		sourceNames = append(sourceNames, source.GetName())
	}

	// 统计成功和失败的数据源数量
	successCount := 0
	failCount := 0

	for _, sourceName := range sourceNames {
		itemCount := resultsStats[sourceName]
		status := "✓"
		if itemCount == 0 {
			status = "✗"
			failCount++
		} else {
			successCount++
		}

		resultStr := fmt.Sprintf("%s %-20s: %d items", status, sourceName, itemCount)
		fmt.Println(resultStr)
		logger.Println(resultStr)
	}

	// 打印汇总信息
	totalItems := 0
	for _, count := range resultsStats {
		totalItems += count
	}

	fmt.Println(separator)
	fmt.Printf("Total: %d sources, %d succeeded, %d failed\n", totalSources, successCount, failCount)
	fmt.Printf("Total items crawled: %d\n", totalItems)
	fmt.Println(separator)

	logger.Println(separator)
	logger.Printf("Total: %d sources, %d succeeded, %d failed\n", totalSources, successCount, failCount)
	logger.Printf("Total items crawled: %d\n", totalItems)
	logger.Println(separator)
}

// logSourceResults 记录数据源的爬取结果到日志
func logSourceResults(sourceName string, items []models.Item, logger *log.Logger) {
	fmt.Printf("\nReceived %d items from %s\n", len(items), sourceName)
	logger.Printf("\nReceived %d items from %s\n", len(items), sourceName)

	// 记录所有结果
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item.Title)
		fmt.Printf("   URL: %s\n", item.URL)
		fmt.Printf("   Category: %s\n", item.Category)
		logger.Printf("%d. %s\n", i+1, item.Title)
		logger.Printf("   URL: %s\n", item.URL)
		logger.Printf("   Category: %s\n", item.Category)
	}
}
