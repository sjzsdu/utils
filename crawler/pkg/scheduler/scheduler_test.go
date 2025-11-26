package scheduler

import (
	"context"
	"testing"
	"time"
)

// mockTask 是一个用于测试的模拟任务
type mockTask struct {
	id       string
	interval int
	executed chan bool
}

// ID 返回任务ID
func (t *mockTask) ID() string {
	return t.id
}

// Execute 执行任务
func (t *mockTask) Execute(ctx context.Context) error {
	t.executed <- true
	return nil
}

// Interval 返回执行间隔
func (t *mockTask) Interval() int {
	return t.interval
}

// TestNewInMemoryScheduler 创建新的内存调度器测试
func TestNewInMemoryScheduler(t *testing.T) {
	s := NewInMemoryScheduler()
	if s == nil {
		t.Fatal("Expected scheduler, got nil")
	}
}

// TestAddTask 添加任务测试
func TestAddTask(t *testing.T) {
	s := NewInMemoryScheduler()
	
	task := &mockTask{
		id:       "test-task",
		interval: 1,
		executed: make(chan bool, 1),
	}
	
	// 添加任务应该成功
	if err := s.AddTask(task); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 再次添加同一个任务应该失败
	if err := s.AddTask(task); err != ErrTaskExists {
		t.Errorf("Expected ErrTaskExists, got %v", err)
	}
}

// TestRemoveTask 移除任务测试
func TestRemoveTask(t *testing.T) {
	s := NewInMemoryScheduler()
	
	// 移除不存在的任务应该失败
	if err := s.RemoveTask("non-existent"); err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
	
	// 添加任务
	task := &mockTask{
		id:       "test-task",
		interval: 1,
		executed: make(chan bool, 1),
	}
	if err := s.AddTask(task); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 移除任务应该成功
	if err := s.RemoveTask("test-task"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 再次移除应该失败
	if err := s.RemoveTask("test-task"); err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

// TestStartAndStop 启动和停止调度器测试
func TestStartAndStop(t *testing.T) {
	s := NewInMemoryScheduler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动调度器
	if err := s.Start(ctx); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 再次启动应该失败
	if err := s.Start(ctx); err != ErrSchedulerRunning {
		t.Errorf("Expected ErrSchedulerRunning, got %v", err)
	}
	
	// 停止调度器
	if err := s.Stop(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 再次停止应该失败
	if err := s.Stop(); err != ErrSchedulerStopped {
		t.Errorf("Expected ErrSchedulerStopped, got %v", err)
	}
}

// TestTaskExecution 任务执行测试
func TestTaskExecution(t *testing.T) {
	s := NewInMemoryScheduler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 创建一个执行间隔为1秒的任务
	executed := make(chan bool, 2)
	task := &mockTask{
		id:       "test-execution",
		interval: 1,
		executed: executed,
	}
	
	// 添加任务
	if err := s.AddTask(task); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 启动调度器
	if err := s.Start(ctx); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 等待任务执行
	select {
	case <-executed:
		// 任务执行成功
	case <-time.After(2 * time.Second):
		t.Fatal("Expected task to execute, but it didn't")
	}
	
	// 停止调度器
	if err := s.Stop(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestGetTask 获取任务测试
func TestGetTask(t *testing.T) {
	s := NewInMemoryScheduler()
	
	task := &mockTask{
		id:       "test-get",
		interval: 1,
		executed: make(chan bool, 1),
	}
	
	// 添加任务
	if err := s.AddTask(task); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 获取任务应该成功
	taskFromScheduler, err := s.GetTask("test-get")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if taskFromScheduler.ID() != task.ID() {
		t.Errorf("Expected task ID %s, got %s", task.ID(), taskFromScheduler.ID())
	}
	
	// 获取不存在的任务应该失败
	_, err = s.GetTask("non-existent")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

// TestListTasks 列出所有任务测试
func TestListTasks(t *testing.T) {
	s := NewInMemoryScheduler()
	
	// 添加多个任务
	task1 := &mockTask{
		id:       "task-1",
		interval: 1,
		executed: make(chan bool, 1),
	}
	
	task2 := &mockTask{
		id:       "task-2",
		interval: 2,
		executed: make(chan bool, 1),
	}
	
	if err := s.AddTask(task1); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if err := s.AddTask(task2); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 列出任务应该返回两个任务
	tasks := s.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
	
	// 移除一个任务
	if err := s.RemoveTask("task-1"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// 列出任务应该返回一个任务
	tasks = s.ListTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
}
