package tasks

import (
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/types"
)

// TaskRegistry 任务注册器
type TaskRegistry struct {
	tasks []types.Task
}

// NewTaskRegistry 创建任务注册器
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{
		tasks: make([]types.Task, 0),
	}
}

// RegisterTask 注册任务
func (r *TaskRegistry) RegisterTask(task types.Task) {
	r.tasks = append(r.tasks, task)
}

// GetAllTasks 获取所有任务
func (r *TaskRegistry) GetAllTasks() []types.Task {
	return r.tasks
}

// RegisterAllTasks 注册所有任务到管理器
func RegisterAllTasks(taskManager types.TaskManager, svc *svc.ServiceContext) {
	fmt.Println("Starting task registration...")
	registry := NewTaskRegistry()

	// 注册示例任务
	fmt.Println("Registering example task...")
	registry.RegisterTask(NewExampleTask(svc))

	// 注册数据清理任务
	fmt.Println("Registering data cleanup task...")
	registry.RegisterTask(NewDataCleanupTask(svc))

	// 注册统计任务
	fmt.Println("Registering stats task...")
	registry.RegisterTask(NewStatsTask(svc))

	// 在这里可以继续注册更多任务...
	// registry.RegisterTask(NewYourCustomTask(svc))

	// 将所有任务注册到管理器
	fmt.Printf("Registering %d tasks to task manager...\n", len(registry.GetAllTasks()))
	for _, task := range registry.GetAllTasks() {
		fmt.Printf("Registering task: %s\n", task.GetName())
		if err := taskManager.RegisterTask(task); err != nil {
			fmt.Printf("Failed to register task %s: %v\n", task.GetName(), err)
			panic(err)
		}
	}
	fmt.Println("All tasks registered to task manager successfully")
}
