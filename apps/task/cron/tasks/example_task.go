package tasks

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// ExampleTask 示例任务
type ExampleTask struct {
	svc *svc.ServiceContext
}

// NewExampleTask 创建示例任务
func NewExampleTask(svc *svc.ServiceContext) *ExampleTask {
	return &ExampleTask{
		svc: svc,
	}
}

// GetName 获取任务名称
func (t *ExampleTask) GetName() string {
	return "example_task"
}

// GetSpec 获取cron表达式 - 每5秒执行一次
func (t *ExampleTask) GetSpec() string {
	return "*/5 * * * * *" // 每5秒执行一次
}

// GetDescription 获取任务描述
func (t *ExampleTask) GetDescription() string {
	return "示例任务，每5秒执行一次，用于演示定时任务功能"
}

// GetTimeout 获取任务超时时间
func (t *ExampleTask) GetTimeout() time.Duration {
	return 30 * time.Second
}

// Execute 执行任务
func (t *ExampleTask) Execute(ctx context.Context) error {
	zLog.Info("Example task started", zap.String("task", t.GetName()))

	// 模拟一些工作
	time.Sleep(2 * time.Second)

	// 这里可以添加具体的业务逻辑
	// 例如：调用其他服务、处理数据、发送通知等

	zLog.Info("Example task completed", zap.String("task", t.GetName()))
	return nil
}
