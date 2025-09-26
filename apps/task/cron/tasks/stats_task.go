package tasks

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// StatsTask 统计任务
type StatsTask struct {
	svc *svc.ServiceContext
}

// NewStatsTask 创建统计任务
func NewStatsTask(svc *svc.ServiceContext) *StatsTask {
	return &StatsTask{
		svc: svc,
	}
}

// GetName 获取任务名称
func (t *StatsTask) GetName() string {
	return "stats_task"
}

// GetSpec 获取cron表达式 - 每小时执行一次
func (t *StatsTask) GetSpec() string {
	return "0 0 * * * *" // 每小时的第0分钟执行
}

// GetDescription 获取任务描述
func (t *StatsTask) GetDescription() string {
	return "统计任务，每小时执行一次，收集系统统计数据"
}

// GetTimeout 获取任务超时时间
func (t *StatsTask) GetTimeout() time.Duration {
	return 5 * time.Minute
}

// Execute 执行任务
func (t *StatsTask) Execute(ctx context.Context) error {
	zLog.Info("Stats task started", zap.String("task", t.GetName()))

	// 这里可以添加具体的统计逻辑
	// 例如：统计用户活跃度、消息发送量、系统性能指标等

	// 模拟统计工作
	time.Sleep(1 * time.Second)

	zLog.Info("Stats task completed", zap.String("task", t.GetName()))
	return nil
}
