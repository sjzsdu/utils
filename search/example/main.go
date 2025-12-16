package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sjzsdu/utils/search"
)

func main() {
	// 从环境变量获取API密钥（如果已设置）
	bingAPIKey := os.Getenv("BING_API_KEY")
	baiduAPIKey := os.Getenv("BAIDU_API_KEY")
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	googleSearchEngineID := os.Getenv("GOOGLE_CSE_ID")

	// 如果环境变量未设置，使用演示值
	if bingAPIKey == "" {
		bingAPIKey = "your-bing-api-key" // 实际使用时会自动从BING_API_KEY环境变量获取
		fmt.Println("注意: BING_API_KEY环境变量未设置，使用演示值")
	}
	if baiduAPIKey == "" {
		baiduAPIKey = "your-baidu-api-key" // 实际使用时会自动从BAIDU_API_KEY环境变量获取
		fmt.Println("注意: BAIDU_API_KEY环境变量未设置，使用演示值")
	}
	if googleAPIKey == "" {
		googleAPIKey = "your-google-api-key" // 实际使用时会自动从GOOGLE_API_KEY环境变量获取
		fmt.Println("注意: GOOGLE_API_KEY环境变量未设置，使用演示值")
	}
	if googleSearchEngineID == "" {
		googleSearchEngineID = "your-google-search-engine-id" // 实际使用时会自动从GOOGLE_CSE_ID环境变量获取
		fmt.Println("注意: GOOGLE_CSE_ID环境变量未设置，使用演示值")
	}

	fmt.Println()

	// 方式1：使用环境变量创建客户端（不提供显式密钥）
	fmt.Println("--- 使用环境变量创建客户端示例 ---")
	envClient := search.NewClient()

	// 注册搜索引擎时不提供显式密钥，将自动从环境变量获取
	envClient.RegisterEngine(search.NewBingSearch(""))       // 自动使用BING_API_KEY
	envClient.RegisterEngine(search.NewBaiduSearch(""))      // 自动使用BAIDU_API_KEY
	envClient.RegisterEngine(search.NewGoogleSearch("", "")) // 自动使用GOOGLE_API_KEY和GOOGLE_CSE_ID

	// 设置默认搜索引擎
	envClient.SetDefaultEngine("bing")

	// 打印已注册的搜索引擎
	fmt.Printf("已注册的搜索引擎: %v\n\n", envClient.ListEngines())

	// 方式2：使用显式API密钥创建客户端
	fmt.Println("--- 使用显式API密钥创建客户端示例 ---")
	client, err := search.NewDefaultClient(
		bingAPIKey,
		baiduAPIKey,
		googleAPIKey,
		googleSearchEngineID,
		search.WithTimeout(20), // 设置超时时间为20秒
	)
	if err != nil {
		log.Fatalf("创建搜索客户端失败: %v", err)
	}

	// 设置默认搜索引擎
	err = client.SetDefaultEngine("bing")
	if err != nil {
		log.Fatalf("设置默认搜索引擎失败: %v", err)
	}

	// 搜索查询
	query := "golang 编程"
	limit := 5

	// 使用默认搜索引擎搜索
	fmt.Printf("\n使用默认搜索引擎 (%s) 搜索: %s\n", client.ListEngines()[0], query)
	results, err := client.Search(context.Background(), query, limit)
	if err != nil {
		log.Printf("搜索失败: %v", err)
	} else {
		printResults(results)
	}

	// 使用指定搜索引擎搜索
	fmt.Printf("\n使用指定搜索引擎 (google) 搜索: %s\n", query)
	results, err = client.SearchWithEngine(context.Background(), "google", query, limit)
	if err != nil {
		log.Printf("搜索失败: %v", err)
	} else {
		printResults(results)
	}

	// 使用选项指定搜索引擎
	fmt.Printf("\n使用选项指定搜索引擎 (baidu) 搜索: %s\n", query)
	results, err = client.Search(context.Background(), query, limit, search.WithEngine("baidu"))
	if err != nil {
		log.Printf("搜索失败: %v", err)
	} else {
		printResults(results)
	}
}

// 打印搜索结果
func printResults(results []search.SearchResult) {
	for i, result := range results {
		fmt.Printf("\n结果 %d:\n", i+1)
		fmt.Printf("标题: %s\n", result.Title)
		fmt.Printf("URL: %s\n", result.URL)
		fmt.Printf("摘要: %s\n", result.Snippet)
	}
}
