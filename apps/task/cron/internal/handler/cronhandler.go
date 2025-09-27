package handler

import (
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/tasks"

	"github.com/zeromicro/go-zero/core/service"
)

// CronHandler 定时任务处理器
type CronHandler struct {
	svc         *svc.ServiceContext
	taskManager types.TaskManager
}

// NewCronHandler 创建定时任务处理器
func NewCronHandler(svc *svc.ServiceContext) *CronHandler {
	fmt.Println("Creating cron handler...")

	// 创建任务管理器
	taskManager := logic.NewTaskManager(
		svc.Config.Cron.WithSeconds,
		100, // 每个任务最多保存100条执行记录
	)

	fmt.Println("Task manager created, registering tasks...")

	// 注册所有任务
	registerAllTasks(taskManager, svc)

	fmt.Println("All tasks registered successfully")

	return &CronHandler{
		svc:         svc,
		taskManager: taskManager,
	}
}

// Services 返回服务列表
func (h *CronHandler) Services() []service.Service {
	// 返回一个简单的服务，用于管理定时任务的生命周期
	return []service.Service{
		&cronService{
			taskManager: h.taskManager,
		},
	}
}

// registerAllTasks 注册所有任务
func registerAllTasks(taskManager types.TaskManager, svc *svc.ServiceContext) {
	// 使用任务注册器注册所有任务
	tasks.RegisterAllTasks(taskManager, svc)
}

// cronService 实现service.Service接口
type cronService struct {
	taskManager types.TaskManager
}

func (s *cronService) Start() {
	fmt.Println("Starting cron service...")
	// 启动任务调度器
	if err := s.taskManager.Start(); err != nil {
		fmt.Printf("Failed to start task manager: %v\n", err)
		panic(err)
	}
	fmt.Println("Cron service started successfully")
}

func (s *cronService) Stop() {
	// 停止任务调度器
	if err := s.taskManager.Stop(); err != nil {
		// 记录错误但不panic
	}
}
