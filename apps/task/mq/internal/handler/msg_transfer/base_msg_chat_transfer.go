package msg_transfer

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

// BaseChatTransfer 实现ConsumeHandler接口
type BaseChatTransfer struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

// NewBaseMsgChatTransfer 实例化一个MsgChatTransfer
func NewBaseMsgChatTransfer(svc *svc.ServiceContext) *BaseChatTransfer {
	return &BaseChatTransfer{
		Logger: logx.WithContext(context.Background()),
		svcCtx: svc,
	}
}

func (m *BaseChatTransfer) MsgChatTransfer(ctx context.Context, data *mq.MsgChatTransfer) error {
	var (
		err error
	)

	fmt.Printf("ws客户端已经收到消息了，准备发送到ws服务端: %+v\n", data)

	switch data.ChatType {
	case constants.SingleChatType:
		//推送至 WebSocket 客户端
		err = m.single(ctx, data)
	case constants.GroupChatType:
		err = m.group(ctx, data)
	}

	return err
}

func (m *BaseChatTransfer) single(ctx context.Context, data *mq.MsgChatTransfer) error {
	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})
}

func (m *BaseChatTransfer) group(ctx context.Context, data *mq.MsgChatTransfer) error {
	res, err := m.svcCtx.Social.GroupUsers(ctx, &socialclient.GroupUsersReq{
		GroupId: data.RecvId,
	})
	if err != nil {
		return err
	}

	data.RecvIdList = make([]string, 0, len(res.List))
	for _, member := range res.List {
		// 跳过发送人
		if member.UserId == data.SendId {
			continue
		}
		data.RecvIdList = append(data.RecvIdList, member.UserId)
	}

	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})
}

// 普通聊天消息（非已读回执）需要把 MsgId 回推给 SENDER，让 sender 前端把 local_ 占位 ID 替换为真实 ID。
// 对已读回执（MsgType=ContentMakeRead）跳过，避免 sender 收到重复的回执。
func (m *BaseChatTransfer) echoToSender(data *mq.MsgChatTransfer) error {
	if data.MsgType == constants.ContentMakeRead || data.ContentType == constants.ContentMakeRead {
		return nil
	}
	if data.SendId == "" || data.MsgId == "" {
		return nil
	}
	// 构造一条专门给发送方的 echo，只带 MsgId/SendTime/ConversationId 等最小字段，
	// 前端据此把 local_ 记录升级到真实 ID。
	echo := *data
	// group 的 RecvIdList 已被上游填充；对 echo 清空，避免 pusher 重复 fan-out
	echo.RecvIdList = nil
	echo.RecvId = data.SendId
	// 给前端一个明确标识：这是发送方回响（而非对方推送）
	echo.ContentType = constants.ContentMsgAck
	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      &echo,
	})
}
