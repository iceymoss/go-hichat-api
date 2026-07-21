// Package notifypush 把 trend-rpc 返回的已生成消息投递到 Kafka，由 task/mq 消费后实时推送。
// 推送是主操作的副作用：单条投递失败仅记录日志，不影响接口返回（与撤回链路一致，最终一致）。
package notifypush

import (
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"go.uber.org/zap"
)

// Publish 把 rpc 返回的 created_messages 逐条投递到 trendNotifyTransfer topic。
func Publish(client mq_client.TrendNotifyTransferClient, messages []*trend.TrendMessageInfo) {
	for _, m := range messages {
		if m == nil {
			continue
		}
		if err := client.Push(&mq.TrendNotifyTransfer{
			ReceiverId:      strconv.FormatUint(m.ReceiverId, 10),
			ActorId:         strconv.FormatUint(m.ActorId, 10),
			MsgType:         int(m.Type),
			TrendId:         m.TrendId,
			CommentId:       m.CommentId,
			ParentCommentId: m.ParentCommentId,
			MessageId:       m.Id,
			Content:         m.Content,
			CreateTime:      m.CreateTime,
		}); err != nil {
			zLog.Error("notifypush.Publish: push failed", zap.Uint64("messageId", m.Id), zap.Error(err))
		}
	}
}
