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
