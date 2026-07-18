package msg_transfer

import (
	"context"
	"encoding/json"
	"time"

	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
)

// CommonNotifyTransfer 公共通知消费者，实现 ConsumeHandler 接口。
// 单一公共 topic（commonNotifyTransfer）：先幂等落库（im notifications 表），再 push.notify 在线推送给接收者。
// 落库与 ChatLogModel 一致直接走 im model；离线由 ws 侧静默丢弃，上线拉列表补偿。
type CommonNotifyTransfer struct {
	*BaseChatTransfer
}

// NewCommonNotifyTransfer 实例化一个 CommonNotifyTransfer
func NewCommonNotifyTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &CommonNotifyTransfer{
		BaseChatTransfer: NewBaseMsgChatTransfer(svc),
	}
}

func (m *CommonNotifyTransfer) Consume(ctx context.Context, key, value string) error {
	var in mq.CommonNotify
	if err := json.Unmarshal([]byte(value), &in); err != nil {
		// 不可重试错误（脏数据）：打日志后丢弃，不阻塞消费组
		logx.WithContext(ctx).Errorf("CommonNotifyTransfer bad message: err=%v", err)
		return nil
	}

	// 防御：自己通知自己 / 空接收者，直接跳过
	if in.ReceiverId == "" || in.ReceiverId == in.ActorId {
		return nil
	}

	createTime := in.CreateTime
	if createTime == 0 {
		createTime = time.Now().Unix()
	}

	// 1. 幂等落库：命中 (receiver_id, notify_type, biz_id) 唯一键则 inserted=false（已处理过）。
	row := &model.Notification{
		ReceiverId: in.ReceiverId,
		NotifyType: in.NotifyType,
		BizId:      in.BizId,
		ActorId:    in.ActorId,
		GroupId:    in.GroupId,
		Title:      in.Title,
		Content:    in.Content,
		Payload:    in.Payload,
		IsRead:     0,
		CreatedAt:  time.Unix(createTime, 0),
	}
	inserted, err := m.svcCtx.NotificationModel.Insert(ctx, row)
	if err != nil {
		// 可重试错误：返回让 kafka 重投
		return err
	}

	// 2. 仅首次落库才推送，避免 kafka 重投造成重复气泡（重复消费无副作用）。
	if !inserted {
		return nil
	}

	if err := m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push.notify",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data: &ws.Notify{
			NotifyType: in.NotifyType,
			ReceiverId: in.ReceiverId,
			ActorId:    in.ActorId,
			BizId:      in.BizId,
			GroupId:    in.GroupId,
			Title:      in.Title,
			Content:    in.Content,
			Payload:    in.Payload,
			CreateTime: createTime,
		},
	}); err != nil {
		logx.WithContext(ctx).Errorf("CommonNotifyTransfer websocket delivery failed: event=%d type=%s receiver=%s err=%v", in.EventId, in.NotifyType, in.ReceiverId, err)
	}
	return nil
}
