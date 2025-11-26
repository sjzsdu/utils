package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrTaskExists 表示任务已存在
	// ErrTaskExists indicates that the task already exists
	ErrTaskExists = errors.New("task already exists")

	// ErrTaskNotFound 表示任务不存在
	// ErrTaskNotFound indicates that the task was not found
	ErrTaskNotFound = errors.New("task not found")

	// ErrSchedulerRunning 表示调度器正在运行
	// ErrSchedulerRunning indicates that the scheduler is running
	ErrSchedulerRunning = errors.New("scheduler is running")

	// ErrSchedulerStopped 表示调度器已停止
	// ErrSchedulerStopped indicates that the scheduler is stopped
	ErrSchedulerStopped = errors.New("scheduler is stopped")
)

// inMemoryScheduler 是 Scheduler 接口的内存实现
// inMemoryScheduler is an in-memory implementation of the Scheduler interface
type inMemoryScheduler struct {
	// tasks 存储所有任务
	// tasks stores all tasks
	tasks map[string]Task

	// taskCtxs 存储每个任务的上下文和取消函数
	// taskCtxs stores the context and cancel function for each task
	taskCtxs map[string]context.CancelFunc

	// mu 保护 tasks 和 taskCtxs 的并发访问
	// mu protects concurrent access to tasks and taskCtxs
	mu sync.RWMutex

	// ctx 调度器的上下文
	// ctx is the scheduler's context
	ctx context.Context

	// cancel 调度器的取消函数
	// cancel is the scheduler's cancel function
	cancel context.CancelFunc

	// wg 用于等待所有任务完成
	// wg is used to wait for all tasks to complete
	wg sync.WaitGroup

	// running 表示调度器是否正在运行
	// running indicates whether the scheduler is running
	running bool
}

// NewInMemoryScheduler 创建一个新的内存调度器
// NewInMemoryScheduler creates a new in-memory scheduler
func NewInMemoryScheduler() Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &inMemoryScheduler{
		tasks:    make(map[string]Task),
		taskCtxs: make(map[string]context.CancelFunc),
		ctx:      ctx,
		cancel:   cancel,
		running:  false,
	}
}

// AddTask 添加任务到调度器
// AddTask adds a task to the scheduler
func (s *inMemoryScheduler) AddTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := task.ID()
	if _, exists := s.tasks[id]; exists {
		return ErrTaskExists
	}

	s.tasks[id] = task

	// 如果调度器正在运行，立即启动该任务
	// If the scheduler is running, start the task immediately
	if s.running {
		s.startTask(task)
	}

	return nil
}

// RemoveTask 从调度器中移除任务
// RemoveTask removes a task from the scheduler
func (s *inMemoryScheduler) RemoveTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	// 取消任务的执行
	// Cancel the task's execution
	if cancel, exists := s.taskCtxs[id]; exists {
		cancel()
		delete(s.taskCtxs, id)
	}

	// 删除任务
	// Delete the task
	delete(s.tasks, id)

	return nil
}

// Start 启动调度器
// Start starts the scheduler
func (s *inMemoryScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrSchedulerRunning
	}

	s.running = true

	// 启动所有任务
	// Start all tasks
	for _, task := range s.tasks {
		s.startTask(task)
	}

	return nil
}

// Stop 停止调度器
// Stop stops the scheduler
func (s *inMemoryScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return ErrSchedulerStopped
	}

	s.running = false

	// 取消所有任务的执行
	// Cancel all tasks' execution
	for _, cancel := range s.taskCtxs {
		cancel()
	}

	// 清空任务上下文
	// Clear task contexts
	s.taskCtxs = make(map[string]context.CancelFunc)

	// 等待所有任务完成
	// Wait for all tasks to complete
	s.wg.Wait()

	return nil
}

// GetTask 获取指定ID的任务
// GetTask gets the task with the specified ID
func (s *inMemoryScheduler) GetTask(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// ListTasks 列出所有任务
// ListTasks lists all tasks
func (s *inMemoryScheduler) ListTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// startTask 启动单个任务
// startTask starts a single task
func (s *inMemoryScheduler) startTask(task Task) {
	id := task.ID()

	// 创建任务的上下文
	// Create the task's context
	taskCtx, cancel := context.WithCancel(s.ctx)
	s.taskCtxs[id] = cancel

	s.wg.Add(1)

	go func() {
		defer func() {
			s.wg.Done()

			s.mu.Lock()
			delete(s.taskCtxs, id)
			s.mu.Unlock()
		}()

		interval := time.Duration(task.Interval()) * time.Second

		// 创建定时器
		// Create a ticker
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// 立即执行一次
		// Execute immediately
		s.executeTask(taskCtx, task)

		for {
			select {
			case <-taskCtx.Done():
				return
			case <-ticker.C:
				s.executeTask(taskCtx, task)
			}
		}
	}()
}

// executeTask 执行单个任务
// executeTask executes a single task
func (s *inMemoryScheduler) executeTask(ctx context.Context, task Task) {
	if err := task.Execute(ctx); err != nil {
		fmt.Printf("Failed to execute task %s: %v\n", task.ID(), err)
	}
}
