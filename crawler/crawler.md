# Go 语言数据爬取包

这是一个基于 Go 语言实现的数据爬取包，遵循开闭原则和基于接口原则，具有良好的扩展性和灵活性。

## 设计原则

- **开闭原则**：对扩展开放，对修改关闭，允许轻松添加新的爬取源
- **基于接口原则**：依赖于抽象而非具体实现，提高代码的灵活性和可测试性
- **单一职责原则**：每个组件只负责一个特定功能
- **依赖倒置原则**：高层模块不依赖低层模块，两者都依赖抽象
- **模块化设计**：将功能划分为独立的子包，便于维护和扩展

## 包结构

```
crawler/
├── cmd/                  # 命令行工具
│   └── crawler/          # 命令行工具入口
├── internal/             # 内部包，不对外暴露
│   ├── cache/            # 缓存实现
│   ├── fetcher/          # HTTP 请求处理
│   └── parser/           # 通用解析工具
├── pkg/                  # 对外暴露的包
│   ├── crawler/          # 核心爬取引擎
│   ├── extractor/        # 数据提取器接口和实现
│   ├── logger/           # 日志工具
│   ├── models/           # 数据模型定义
│   └── scheduler/        # 爬取任务调度器
├── sources/              # 各种数据源的实现
│   ├── github/           # GitHub 数据源
│   ├── news/             # 新闻网站数据源
│   └── source.go         # 数据源注册机制
```

## 核心接口

### Source 接口
定义了数据源的基本行为，包括获取名称、URL、内容、解析内容和获取爬取间隔。

### Engine 接口
定义了爬取引擎的基本行为，包括注册数据源、启动/停止引擎、获取数据和订阅更新。

### Extractor 接口
定义了数据提取的基本行为，包括提取标题、内容、链接、图片和时间。

### Cache 接口
定义了缓存的基本行为，包括获取、设置、删除和清空缓存。

## UML 类图

```mermaid
classDiagram
    direction LR
    
    %% 核心接口
    class Source {
        +GetName() string
        +GetURL() string
        +Fetch(ctx context.Context) ([]byte, error)
        +Parse(content []byte) ([]Item, error)
        +GetInterval() int
    }
    
    class Engine {
        +RegisterSource(source Source)
        +Start(ctx context.Context)
        +Stop()
        +GetItems(sourceName string) ([]Item, error)
        +Subscribe(sourceName string, ch chan<- []Item)
    }
    
    class Extractor {
        +ExtractTitle(content []byte) (string, error)
        +ExtractContent(content []byte) (string, error)
        +ExtractLinks(content []byte) ([]string, error)
        +ExtractImages(content []byte) ([]string, error)
        +ExtractTime(content []byte) (time.Time, error)
    }
    
    class Cache {
        +Get(key string) ([]Item, error)
        +Set(key string, items []Item, expiration time.Duration) error
        +Delete(key string) error
        +Clear() error
        +Close() error
    }
    
    %% 数据模型
    class Item {
        -ID string
        -Title string
        -URL string
        -Content string
        -Source string
        -Category string
        -Images []string
        -PublishedAt time.Time
        -CreatedAt time.Time
        -UpdatedAt time.Time
    }
    
    %% 内部组件
    class Fetcher {
        +Get(ctx context.Context, url string) ([]byte, error)
    }
    
    class Parser {
        +ParseElement(content []byte, selector string) (*goquery.Selection, error)
        +ExtractText(content []byte, selector string) (string, error)
        +ExtractAttr(content []byte, selector string, attr string) (string, error)
    }
    
    class Scheduler {
        +AddTask(task func())
        +Start()
        +Stop()
    }
    
    %% 具体实现
    class GitHubSource {
        +GetName() string
        +GetURL() string
        +Fetch(ctx context.Context) ([]byte, error)
        +Parse(content []byte) ([]Item, error)
        +GetInterval() int
    }
    
    class NewsSource {
        +GetName() string
        +GetURL() string
        +Fetch(ctx context.Context) ([]byte, error)
        +Parse(content []byte) ([]Item, error)
        +GetInterval() int
    }
    
    class MemoryCache {
        +Get(key string) ([]Item, error)
        +Set(key string, items []Item, expiration time.Duration) error
        +Delete(key string) error
        +Clear() error
        +Close() error
    }
    
    class DefaultEngine {
        +RegisterSource(source Source)
        +Start(ctx context.Context)
        +Stop()
        +GetItems(sourceName string) ([]Item, error)
        +Subscribe(sourceName string, ch chan<- []Item)
    }
    
    %% 关系
    Engine --> Cache
    Engine --> Source
    Engine --> Scheduler
    
    GitHubSource --> Source
    NewsSource --> Source
    MemoryCache --> Cache
    DefaultEngine --> Engine
    
    Source --> Fetcher
    Source --> Parser
    Source --> Item
    
    Parser --> Extractor
    
    Engine --> Item
    Extractor --> Item
    Cache --> Item
    
    %% 数据源注册
    class SourceRegistry {
        +Register(source Source)
        +List() []Source
        +Get(name string) (Source, error)
    }
    
    SourceRegistry --> Source
    Engine --> SourceRegistry
```

## 时序图

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Engine as 爬取引擎
    participant Scheduler as 调度器
    participant Source as 数据源
    participant Fetcher as HTTP Fetcher
    participant Parser as HTML解析器
    participant Cache as 缓存
    participant Subscriber as 订阅者
    
    %% 初始化阶段
    Client->>Cache: 创建内存缓存
    Client->>Engine: 创建爬取引擎
    Client->>SourceRegistry: 获取数据源注册表
    SourceRegistry->>Engine: 注册所有数据源
    
    %% 启动阶段
    Client->>Engine: 启动引擎
    Engine->>Scheduler: 启动调度器
    
    %% 爬取周期
    loop 按数据源间隔循环
        Scheduler->>Engine: 触发爬取任务
        Engine->>Source: 获取数据源
        Engine->>Cache: 检查缓存
        alt 缓存未命中或已过期
            Engine->>Source: 调用Fetch方法
            Source->>Fetcher: 发送HTTP请求
            Fetcher-->>Source: 返回HTML内容
            Source->>Parser: 解析HTML
            Parser-->>Source: 返回解析结果
            Source-->>Engine: 返回结构化Item数据
            Engine->>Cache: 缓存Item数据
        else 缓存命中
            Cache-->>Engine: 返回缓存的Item数据
        end
        
        %% 通知订阅者
        Engine->>Subscriber: 发送Item数据
    end
    
    %% 停止阶段
    Client->>Engine: 停止引擎
    Engine->>Scheduler: 停止调度器
    Engine->>Cache: 关闭缓存
    
    %% 订阅更新
    Client->>Engine: 订阅数据源更新
    Engine->>Subscriber: 建立订阅通道
```

## 系统架构图

```mermaid
flowchart TD
    %% 客户端层
    Client["客户端应用"]
    
    %% 核心引擎层
    Engine["爬取引擎"]
    Scheduler["调度器"]
    SourceRegistry["数据源注册表"]
    
    %% 数据源层
    Sources["数据源集合"]
    GitHub["GitHub数据源"]
    News["新闻数据源"]
    Custom["自定义数据源"]
    
    %% 内部服务层
    Fetcher["HTTP Fetcher"]
    Parser["HTML解析器"]
    Cache["内存缓存"]
    
    %% 订阅层
    Subscribers["订阅者集合"]
    Subscriber1["订阅者1"]
    Subscriber2["订阅者2"]
    
    %% 数据流
    Client --> Engine
    Client --> SourceRegistry
    
    Engine --> Scheduler
    Engine --> SourceRegistry
    Engine --> Sources
    Engine --> Cache
    Engine --> Subscribers
    
    SourceRegistry --> Sources
    
    Sources --> GitHub
    Sources --> News
    Sources --> Custom
    
    GitHub --> Fetcher
    News --> Fetcher
    Custom --> Fetcher
    
    GitHub --> Parser
    News --> Parser
    Custom --> Parser
    
    Subscribers --> Subscriber1
    Subscribers --> Subscriber2
    
    %% 组件间依赖关系
    classDef core fill:#f9d5e5,stroke:#333,stroke-width:2px
    classDef component fill:#e5f9d5,stroke:#333,stroke-width:2px
    classDef external fill:#d5e5f9,stroke:#333,stroke-width:2px
    
    class Engine,Scheduler,SourceRegistry core
    class Fetcher,Parser,Cache component
    class GitHub,News,Custom,Subscriber1,Subscriber2,Client external
```

## 使用示例

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sjzsdu/utils/crawler/internal/cache"
	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/sources"
	_ "github.com/sjzsdu/utils/crawler/sources/github"
	_ "github.com/sjzsdu/utils/crawler/sources/news"
)

func main() {
	// 创建内存缓存
	memCache := cache.NewMemoryCache(1 * time.Hour)
	defer memCache.Close()
	
	// 创建爬取引擎
	engine := crawler.NewEngine(memCache)
	
	// 注册所有数据源
	for _, source := range sources.GetRegistry().List() {
		engine.RegisterSource(source)
	}
	
	// 启动爬取引擎
	engine.Start(context.Background())
	defer engine.Stop()
	
	// 订阅更新
	githubChan := make(chan []models.Item, 10)
	engine.Subscribe("github", githubChan)
	
	// 处理爬取结果
	for items := range githubChan {
		fmt.Printf("Received %d items from GitHub\n", len(items))
	}
}
```

## 添加新数据源

要添加新的数据源，只需实现 `Source` 接口并注册到注册表中：

```go
package mysource

import (
	"context"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/crawler"
	"github.com/sjzsdu/utils/crawler/pkg/models"
	"github.com/sjzsdu/utils/crawler/sources"
)

// mySource 是自定义数据源的实现
type mySource struct {
	// 实现 Source 接口的方法
}

func init() {
	// 注册数据源
	sources.GetRegistry().Register(&mySource{})
}
```

## 运行示例

```bash
go run cmd/crawler/main.go
```

## 运行测试

```bash
go test ./...
```

## 性能优化

1. **并发爬取**：使用 goroutine 并发爬取多个数据源
2. **智能缓存**：根据数据源的更新频率调整缓存过期时间
3. **失败重试**：实现指数退避重试机制，提高爬取成功率
4. **限速控制**：对每个数据源实施独立的限速策略
5. **连接池**：使用 HTTP 连接池，减少连接建立和关闭的开销
6. **增量爬取**：只爬取新增或更新的数据

## 监控与日志

1. **日志记录**：使用结构化日志，记录爬取过程中的关键事件和错误
2. **指标监控**：暴露 Prometheus 指标，包括爬取成功率、响应时间、数据量等
3. **健康检查**：提供健康检查接口，便于监控系统集成
4. **告警机制**：当爬取失败率超过阈值时，发送告警通知

## 许可证

MIT
