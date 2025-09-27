package types

import (
	"context"
	"time"
)

// Task 定时任务接口
type Task interface {
	// GetName 获取任务名称
	GetName() string
	// GetSpec 获取cron表达式
	GetSpec() string
	// GetDescription 获取任务描述
	GetDescription() string
	// Execute 执行任务
	Execute(ctx context.Context) error
	// GetTimeout 获取任务超时时间
	GetTimeout() time.Duration
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskName  string        `json:"task_name"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Message   string        `json:"message,omitempty"`
}

// TaskManager 任务管理器接口
type TaskManager interface {
	// RegisterTask 注册任务
	RegisterTask(task Task) error
	// UnregisterTask 取消注册任务
	UnregisterTask(taskName string) error
	// Start 启动任务调度器
	Start() error
	// Stop 停止任务调度器
	Stop() error
	// GetTaskStatus 获取任务状态
	GetTaskStatus(taskName string) (bool, error)
	// GetTaskResults 获取任务执行结果
	GetTaskResults(taskName string, limit int) ([]TaskResult, error)
}
