package scheduler

import (
	"context"
)

// Task 定义了一个调度任务
// Task defines a scheduled task
type Task interface {
	// ID 返回任务的唯一标识符
	// ID returns the unique identifier of the task
	ID() string

	// Execute 执行任务
	// Execute executes the task
	Execute(ctx context.Context) error

	// Interval 返回任务的执行间隔（秒）
	// Interval returns the execution interval of the task in seconds
	Interval() int
}

// Scheduler 定义了调度器的接口
// Scheduler defines the scheduler interface
type Scheduler interface {
	// AddTask 添加任务到调度器
	// AddTask adds a task to the scheduler
	AddTask(task Task) error

	// RemoveTask 从调度器中移除任务
	// RemoveTask removes a task from the scheduler
	RemoveTask(id string) error

	// Start 启动调度器
	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop 停止调度器
	// Stop stops the scheduler
	Stop() error

	// GetTask 获取指定ID的任务
	// GetTask gets the task with the specified ID
	GetTask(id string) (Task, error)

	// ListTasks 列出所有任务
	// ListTasks lists all tasks
	ListTasks() []Task
}
