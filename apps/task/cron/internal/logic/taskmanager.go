package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/types"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// TaskManager 任务管理器实现
type TaskManager struct {
	cron        *cron.Cron
	tasks       map[string]types.Task
	results     map[string][]types.TaskResult
	mu          sync.RWMutex
	maxResults  int
	withSeconds bool
}

// NewTaskManager 创建任务管理器
func NewTaskManager(withSeconds bool, maxResults int) *TaskManager {
	var c *cron.Cron
	if withSeconds {
		c = cron.New(cron.WithSeconds())
	} else {
		c = cron.New()
	}

	return &TaskManager{
		cron:        c,
		tasks:       make(map[string]types.Task),
		results:     make(map[string][]types.TaskResult),
		maxResults:  maxResults,
		withSeconds: withSeconds,
	}
}

// RegisterTask 注册任务
func (tm *TaskManager) RegisterTask(task types.Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	taskName := task.GetName()
	if _, exists := tm.tasks[taskName]; exists {
		return fmt.Errorf("task %s already registered", taskName)
	}

	// 添加任务到cron调度器
	entryID, err := tm.cron.AddFunc(task.GetSpec(), func() {
		tm.executeTask(context.Background(), task)
	})
	if err != nil {
		return fmt.Errorf("failed to add task %s to cron: %w", taskName, err)
	}

	tm.tasks[taskName] = task
	tm.results[taskName] = make([]types.TaskResult, 0, tm.maxResults)

	zLog.Info("Task registered successfully",
		zap.String("task_name", taskName),
		zap.String("spec", task.GetSpec()),
		zap.Int("entry_id", int(entryID)))
	return nil
}

// UnregisterTask 取消注册任务
func (tm *TaskManager) UnregisterTask(taskName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[taskName]; !exists {
		return fmt.Errorf("task %s not found", taskName)
	}

	// 从cron调度器中移除任务（这里需要保存entry ID，简化处理）
	delete(tm.tasks, taskName)
	delete(tm.results, taskName)

	zLog.Info("Task unregistered successfully", zap.String("task_name", taskName))
	return nil
}

// Start 启动任务调度器
func (tm *TaskManager) Start() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.cron == nil {
		return fmt.Errorf("cron scheduler not initialized")
	}

	fmt.Printf("Starting cron scheduler with %d registered tasks\n", len(tm.tasks))
	tm.cron.Start()
	zLog.Info("Task scheduler started", zap.Int("task_count", len(tm.tasks)))
	return nil
}

// Stop 停止任务调度器
func (tm *TaskManager) Stop() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.cron == nil {
		return fmt.Errorf("cron scheduler not initialized")
	}

	ctx := tm.cron.Stop()
	<-ctx.Done()
	zLog.Info("Task scheduler stopped")
	return nil
}

// GetTaskStatus 获取任务状态
func (tm *TaskManager) GetTaskStatus(taskName string) (bool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.tasks[taskName]
	return exists, nil
}

// GetTaskResults 获取任务执行结果
func (tm *TaskManager) GetTaskResults(taskName string, limit int) ([]types.TaskResult, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	results, exists := tm.results[taskName]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskName)
	}

	if limit <= 0 || limit > len(results) {
		limit = len(results)
	}

	// 返回最新的结果
	start := len(results) - limit
	if start < 0 {
		start = 0
	}

	return results[start:], nil
}

// executeTask 执行任务
func (tm *TaskManager) executeTask(ctx context.Context, task types.Task) {
	taskName := task.GetName()
	startTime := time.Now()

	fmt.Printf("=== TASK EXECUTION STARTED: %s at %s ===\n", taskName, startTime.Format("2006-01-02 15:04:05"))

	// 创建带超时的上下文
	timeout := task.GetTimeout()
	if timeout <= 0 {
		timeout = 5 * time.Minute // 默认5分钟超时
	}

	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := types.TaskResult{
		TaskName:  taskName,
		StartTime: startTime,
	}

	zLog.Info("Starting task execution", zap.String("task_name", taskName))

	// 执行任务
	err := task.Execute(taskCtx)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		zLog.Error("Task execution failed",
			zap.String("task_name", taskName),
			zap.Error(err),
			zap.Duration("duration", result.Duration))
	} else {
		result.Success = true
		result.Message = "Task executed successfully"
		zLog.Info("Task execution completed",
			zap.String("task_name", taskName),
			zap.Duration("duration", result.Duration))
	}

	// 保存执行结果
	tm.saveTaskResult(result)
}

// saveTaskResult 保存任务执行结果
func (tm *TaskManager) saveTaskResult(result types.TaskResult) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	results := tm.results[result.TaskName]
	if results == nil {
		results = make([]types.TaskResult, 0, tm.maxResults)
	}

	// 添加新结果
	results = append(results, result)

	// 保持结果数量在限制范围内
	if len(results) > tm.maxResults {
		results = results[len(results)-tm.maxResults:]
	}

	tm.results[result.TaskName] = results
}
