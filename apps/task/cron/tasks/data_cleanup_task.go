package tasks

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// DataCleanupTask 数据清理任务
type DataCleanupTask struct {
	svc *svc.ServiceContext
}

// NewDataCleanupTask 创建数据清理任务
func NewDataCleanupTask(svc *svc.ServiceContext) *DataCleanupTask {
	return &DataCleanupTask{
		svc: svc,
	}
}

// GetName 获取任务名称
func (t *DataCleanupTask) GetName() string {
	return "data_cleanup_task"
}

// GetSpec 获取cron表达式 - 每天凌晨2点执行
func (t *DataCleanupTask) GetSpec() string {
	return "0 0 2 * * *" // 每天凌晨2点执行
}

// GetDescription 获取任务描述
func (t *DataCleanupTask) GetDescription() string {
	return "数据清理任务，每天凌晨2点执行，清理过期数据"
}

// GetTimeout 获取任务超时时间
func (t *DataCleanupTask) GetTimeout() time.Duration {
	return 10 * time.Minute
}

// Execute 执行任务
func (t *DataCleanupTask) Execute(ctx context.Context) error {
	zLog.Info("Data cleanup task started", zap.String("task", t.GetName()))

	// 清理30天前的聊天记录
	cutoffTime := time.Now().AddDate(0, 0, -30)

	// 这里可以添加具体的清理逻辑
	// 例如：删除过期的聊天记录、清理临时文件等

	zLog.Info("Data cleanup task completed",
		zap.String("task", t.GetName()),
		zap.Time("cutoff_time", cutoffTime))

	return nil
}
