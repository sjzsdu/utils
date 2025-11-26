package crawler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
	"github.com/sjzsdu/utils/crawler/pkg/scheduler"
)

// engineImpl 是 Engine 接口的实现
type engineImpl struct {
	// 注册的数据源映射
	sources map[string]Source

	// 订阅者映射
	subscribers map[string][]chan<- []models.Item

	// 缓存
	cache Cache

	// 上下文
	ctx context.Context

	// 取消函数
	cancel context.CancelFunc

	// 互斥锁
	mu sync.RWMutex

	// 运行状态
	running bool

	// 调度器
	scheduler scheduler.Scheduler
}

// crawlTask 实现了 scheduler.Task 接口，用于爬取数据源
type crawlTask struct {
	source Source
	engine *engineImpl
}

// ID 返回任务的唯一标识符
func (t *crawlTask) ID() string {
	return t.source.GetName()
}

// Execute 执行爬取任务
func (t *crawlTask) Execute(ctx context.Context) error {
	t.engine.fetchAndProcess(t.source)
	return nil
}

// Interval 返回任务的执行间隔
func (t *crawlTask) Interval() int {
	return t.source.GetInterval()
}

// NewEngine 创建一个新的爬取引擎实例
func NewEngine(cache Cache) Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &engineImpl{
		sources:     make(map[string]Source),
		subscribers: make(map[string][]chan<- []models.Item),
		cache:       cache,
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
		scheduler:   scheduler.NewInMemoryScheduler(),
	}
}

// RegisterSource 注册数据源
func (e *engineImpl) RegisterSource(source Source) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	name := source.GetName()
	if _, exists := e.sources[name]; exists {
		return fmt.Errorf("source %s already registered", name)
	}

	e.sources[name] = source

	// 如果引擎正在运行，将任务添加到调度器
	if e.running {
		task := &crawlTask{
			source: source,
			engine: e,
		}
		return e.scheduler.AddTask(task)
	}

	return nil
}

// UnregisterSource 注销数据源
func (e *engineImpl) UnregisterSource(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.sources[name]; !exists {
		return fmt.Errorf("source %s not found", name)
	}

	delete(e.sources, name)

	// 如果引擎正在运行，从调度器中移除任务
	if e.running {
		return e.scheduler.RemoveTask(name)
	}

	return nil
}

// Start 启动爬取引擎
func (e *engineImpl) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = true
	e.mu.Unlock()

	// 将所有数据源添加到调度器
	e.mu.RLock()
	for _, source := range e.sources {
		task := &crawlTask{
			source: source,
			engine: e,
		}
		if err := e.scheduler.AddTask(task); err != nil {
			fmt.Printf("Failed to add task for source %s: %v\n", source.GetName(), err)
			continue
		}
	}
	e.mu.RUnlock()

	// 启动调度器
	return e.scheduler.Start(ctx)
}

// Stop 停止爬取引擎
func (e *engineImpl) Stop() error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = false
	e.mu.Unlock()

	// 停止调度器
	if err := e.scheduler.Stop(); err != nil {
		return err
	}

	// 取消上下文
	e.cancel()

	return nil
}

// FetchItem 获取指定数据源的最新数据
func (e *engineImpl) FetchItem(ctx context.Context, sourceName string) ([]models.Item, error) {
	e.mu.RLock()
	source, exists := e.sources[sourceName]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	// 先尝试从缓存获取
	items, err := e.cache.Get(sourceName)
	if err == nil && len(items) > 0 {
		return items, nil
	}

	// 缓存未命中，直接爬取
	content, err := source.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	items, err = source.Parse(content)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	e.cache.Set(sourceName, items, time.Duration(source.GetInterval())*time.Second)

	// 通知订阅者
	e.notifySubscribers(sourceName, items)

	return items, nil
}

// Subscribe 订阅数据源的更新
func (e *engineImpl) Subscribe(sourceName string, ch chan<- []models.Item) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.sources[sourceName]; !exists {
		return fmt.Errorf("source %s not found", sourceName)
	}

	e.subscribers[sourceName] = append(e.subscribers[sourceName], ch)
	return nil
}

// Unsubscribe 取消订阅
func (e *engineImpl) Unsubscribe(sourceName string, ch chan<- []models.Item) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if subscribers, exists := e.subscribers[sourceName]; exists {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				e.subscribers[sourceName] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}

	return nil
}

// fetchAndProcess 获取并处理数据源
func (e *engineImpl) fetchAndProcess(source Source) {
	content, err := source.Fetch(e.ctx)
	if err != nil {
		fmt.Printf("Failed to fetch source %s: %v\n", source.GetName(), err)
		return
	}

	items, err := source.Parse(content)
	if err != nil {
		fmt.Printf("Failed to parse source %s: %v\n", source.GetName(), err)
		return
	}

	name := source.GetName()

	// 更新缓存
	e.cache.Set(name, items, time.Duration(source.GetInterval())*time.Second)

	// 通知订阅者
	e.notifySubscribers(name, items)
}

// notifySubscribers 通知订阅者
func (e *engineImpl) notifySubscribers(sourceName string, items []models.Item) {
	e.mu.RLock()
	subscribers, exists := e.subscribers[sourceName]
	e.mu.RUnlock()

	if !exists {
		return
	}

	// 通知所有订阅者
	for _, ch := range subscribers {
		select {
		case ch <- items:
		default:
			// 如果通道已满，跳过本次通知
			fmt.Printf("Subscriber channel for source %s is full, skipping notification\n", sourceName)
		}
	}
}
