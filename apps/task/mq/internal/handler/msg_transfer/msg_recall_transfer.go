package msg_transfer

import (
	"context"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
)

// MsgRecallTransfer 撤回事件消费者，实现 ConsumeHandler 接口。
// 撤回走独立 topic（msgRecallTransfer），与普通聊天消息解耦；
// DB 已由 im-rpc 更新，这里不落库，仅把撤回翻译成 ContentRecall 控制帧扇出给会话各端。
type MsgRecallTransfer struct {
	*BaseChatTransfer
}

// NewMsgRecallTransfer 实例化一个 MsgRecallTransfer
func NewMsgRecallTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &MsgRecallTransfer{
		BaseChatTransfer: NewBaseMsgChatTransfer(svc),
	}
}

func (m *MsgRecallTransfer) Consume(ctx context.Context, key, value string) error {
	var in mq.MsgRecallTransfer
	if err := json.Unmarshal([]byte(value), &in); err != nil {
		return err
	}

	// 翻译成通用推送帧：保持 ContentRecall（mType=9），前端协议与原先逐字段一致。
	data := &mq.MsgChatTransfer{
		ChatType:       in.ChatType,
		ConversationId: in.ConversationId,
		SendId:         in.SendId,
		RecvId:         in.RecvId,
		MsgId:          in.MsgId,
		MsgType:        constants.ContentRecall,
		ContentType:    constants.ContentRecall,
		RecalledBy:     in.RecalledBy,
	}

	// 先按会话成员扇出（group 跳过原发送者 / single 推给对端），
	// 再单独补推给原发送者本人（含其它在线设备）。撤回幂等，重复推送无副作用。
	if err := m.MsgChatTransfer(ctx, data); err != nil {
		return err
	}
	if data.SendId == "" {
		return nil
	}
	// 补推给原发送者本人：走 single 路径（pusher 会 push 到 RecvId），不改 ContentType。
	echo := *data
	echo.RecvIdList = nil
	echo.RecvId = data.SendId
	echo.ChatType = constants.SingleChatType
	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      &echo,
	})
}
