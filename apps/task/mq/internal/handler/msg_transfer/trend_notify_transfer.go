package msg_transfer

import (
	"context"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
)

// TrendNotifyTransfer 动态消息通知消费者，实现 ConsumeHandler 接口。
// 走独立 topic（trendNotifyTransfer），与聊天/撤回解耦；DB 已由 trend-rpc 落库，
// 这里不落库，仅把通知翻译成 push.trend 帧单推给接收者。
type TrendNotifyTransfer struct {
	*BaseChatTransfer
}

// NewTrendNotifyTransfer 实例化一个 TrendNotifyTransfer
func NewTrendNotifyTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &TrendNotifyTransfer{
		BaseChatTransfer: NewBaseMsgChatTransfer(svc),
	}
}

func (m *TrendNotifyTransfer) Consume(ctx context.Context, key, value string) error {
	var in mq.TrendNotifyTransfer
	if err := json.Unmarshal([]byte(value), &in); err != nil {
		return err
	}

	// 翻译成 push.trend 推送帧，单推给接收者（离线由 ws 侧静默丢弃，上线拉列表补偿）。
	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push.trend",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data: &ws.TrendNotify{
			ReceiverId:      in.ReceiverId,
			ActorId:         in.ActorId,
			MsgType:         in.MsgType,
			TrendId:         in.TrendId,
			CommentId:       in.CommentId,
			ParentCommentId: in.ParentCommentId,
			MessageId:       in.MessageId,
			Content:         in.Content,
			CreateTime:      in.CreateTime,
		},
	})
}
