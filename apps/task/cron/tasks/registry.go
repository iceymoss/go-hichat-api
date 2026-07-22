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

	if svc.Config.Cron.InvitationExpirationSpec != "" {
		fmt.Println("Registering group invitation expiration task...")
		registry.RegisterTask(NewGroupInvitationExpirationTask(svc))
	}

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
