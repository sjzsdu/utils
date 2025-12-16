# Search Package

A flexible and unified search package for Go that supports Bing, Baidu, and Google search engines.

## Features

- Unified interface for multiple search engines
- Support for Bing, Baidu, and Google search APIs
- Flexible configuration using option pattern
- Environment variable support for API keys
- Easy extensibility to add new search engines

## Installation

```bash
go get -u github.com/sjzsdu/utils/search
```

## API Key Setup

### Bing Search API

1. Go to the [Azure Portal](https://portal.azure.com/)
2. Sign in or create an Azure account
3. Create a new "Bing Search v7" resource
4. Navigate to the "Keys and Endpoint" section
5. Copy your API key

Environment variable: `BING_API_KEY`

### Baidu Qianfan AI Search API

1. Go to the [Baidu AI Cloud Console](https://console.bce.baidu.com/)
2. Sign in or create a Baidu AI Cloud account
3. Navigate to "Qianfan AI Platform" > "API Key Management"
4. Create a new application or use an existing one
5. Copy your API key

Environment variable: `BAIDU_API_KEY`

### Google Custom Search API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Sign in or create a Google Cloud account
3. Create a new project or select an existing one
4. Enable the "Custom Search JSON API"
5. Navigate to "Credentials" and create an API key
6. Go to the [Google Custom Search Engine](https://programmablesearchengine.google.com/controlpanel/all) page
7. Create a new search engine and get the Search Engine ID

Environment variables:
- `GOOGLE_API_KEY` - Your Google API key
- `GOOGLE_CSE_ID` - Your Google Custom Search Engine ID

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yourusername/search"
)

func main() {
	// Set environment variables (or set them in your shell)
	nil

	// Create a client
	client := search.NewClient()

	// Register engines with default configuration (using environment variables)
	client.RegisterEngine(search.NewBingSearch(""))
	client.RegisterEngine(search.NewBaiduSearch(""))
	client.RegisterEngine(search.NewGoogleSearch("", ""))

	// Set default engine
	client.SetDefaultEngine("bing")

	// Perform a search
	results, err := client.Search(context.Background(), "Go programming", 10)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		os.Exit(1)
	}

	// Display results
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Title)
		fmt.Printf("   URL: %s\n", result.URL)
		fmt.Printf("   Snippet: %s\n\n", result.Snippet)
	}
}
```

### Manual Configuration

```go
// Register engines with manual API keys
client.RegisterEngine(search.NewBingSearch("your-bing-api-key"))
client.RegisterEngine(search.NewBaiduSearch("your-baidu-api-key"))
client.RegisterEngine(search.NewGoogleSearch("your-google-api-key", "your-google-cse-id"))
```

### With Custom Options

```go
// Register engine with custom timeout and headers
client.RegisterEngine(
	search.NewBingSearch(
		"your-bing-api-key",
		search.WithTimeout(30),
		search.WithHeaders(map[string]string{
			"User-Agent": "Your-App/1.0",
		}),
	),
)
```

## Environment Variables

The search package will automatically use these environment variables if no API key is provided:

- `BING_API_KEY` - Bing Search API key
- `BAIDU_API_KEY` - Baidu Qianfan AI Search API key
- `GOOGLE_API_KEY` - Google Custom Search API key
- `GOOGLE_CSE_ID` - Google Custom Search Engine ID

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
