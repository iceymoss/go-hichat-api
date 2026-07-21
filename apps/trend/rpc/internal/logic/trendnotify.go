package logic

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/notify"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"go.uber.org/zap"
)

// trendMsgContentMax 消息内容快照最大字符数（列宽 500，这里再保守截断）
const trendMsgContentMax = 200

// persistTrendMessages 把一次互动事件翻译成消息集合并落库，返回已创建的消息（供 api 层投 Kafka 推送）。
// 单条落库失败不阻塞其它：通知是主操作的副作用，失败仅记录日志，不回滚（与撤回链路一致）。
func persistTrendMessages(ctx context.Context, svcCtx *svc.ServiceContext, ev notify.TrendNotifyEvent) []*trend.TrendMessageInfo {
	ev.Content = truncateContent(ev.Content)
	specs := notify.BuildTrendMessages(ev)
	out := make([]*trend.TrendMessageInfo, 0, len(specs))
	now := time.Now()
	for _, s := range specs {
		m := &models.TrendMessage{
			ReceiverId:      s.ReceiverId,
			ActorId:         s.ActorId,
			Type:            uint64(s.Type),
			TrendId:         s.TrendId,
			CommentId:       s.CommentId,
			ParentCommentId: s.ParentCommentId,
			Content:         s.Content,
			IsRead:          constants.TrendMsgUnread,
			State:           constants.TrendMsgStateNormal,
			CreateTime:      now,
		}
		id, err := svcCtx.TrendMessage.CreateMsg(ctx, m)
		if err != nil {
			zLog.Error("persistTrendMessages: create trend message failed",
				zap.Uint64("receiver", s.ReceiverId), zap.Int("type", int(s.Type)), zap.Error(err))
			continue
		}
		out = append(out, &trend.TrendMessageInfo{
			Id:              id,
			ReceiverId:      s.ReceiverId,
			ActorId:         s.ActorId,
			Type:            int32(s.Type),
			TrendId:         s.TrendId,
			CommentId:       s.CommentId,
			ParentCommentId: s.ParentCommentId,
			Content:         s.Content,
			IsRead:          false,
			CreateTime:      now.Unix(),
		})
	}
	return out
}

// truncateContent 按字符截断内容快照，避免超列宽 / 推送体过大。
func truncateContent(s string) string {
	r := []rune(s)
	if len(r) <= trendMsgContentMax {
		return s
	}
	return string(r[:trendMsgContentMax])
}
