package msg_transfer

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"
	"github.com/iceymoss/go-hichat-api/pkg/wuid"

	"github.com/zeromicro/go-queue/kq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RelationChangeTransfer 关系变更事件消费者：维护关系缓存（版本门）+ 推送会话失效帧。
// 与聊天/撤回/动态链路解耦，走独立 topic relationChangeTransfer。幂等：版本门 + 重复推送对客户端无副作用。
type RelationChangeTransfer struct {
	*BaseChatTransfer
}

var relationConsumerStopping atomic.Bool

func init() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() { <-shutdown; relationConsumerStopping.Store(true) }()
}

func NewRelationChangeTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &RelationChangeTransfer{
		BaseChatTransfer: NewBaseMsgChatTransfer(svc),
	}
}

func (m *RelationChangeTransfer) Consume(ctx context.Context, key, value string) error {
	if relationConsumerStopping.Load() {
		return context.Canceled
	}
	var in mq.RelationChangeTransfer
	if err := json.Unmarshal([]byte(value), &in); err != nil {
		zLog.Error("RelationChange.Unmarshal", zap.String("value", value), zap.Error(err))
		return nil
	}
	if in.EventType == constants.RelationEventGroupMemberAdded || in.EventType == constants.RelationEventGroupMemberRemoved {
		if !validGroupRelationEvent(&in) {
			zLog.Error("RelationChange.poison", zap.Any("event", in))
			return nil
		}
	}
	if err := m.retryReliableState(ctx, &in); err != nil {
		return err
	}

	// 1) 维护关系缓存（best-effort，版本门保证幂等/防乱序；失败不阻断推送）
	m.applyCache(ctx, &in)

	// 2) 冻结/解冻被移出成员的群历史（被踢只能看被踢前的消息；重新入群解冻）
	// 3) 推送会话失效帧给受影响用户（客户端禁用输入闭环）
	m.pushRelationChanged(&in)
	return nil
}

func (m *RelationChangeTransfer) retryReliableState(ctx context.Context, in *mq.RelationChangeTransfer) error {
	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	delay := time.Second
	for {
		var err error
		switch in.EventType {
		case constants.RelationEventGroupMemberAdded:
			err = ensureGroupMemberConversation(shutdownCtx, m.svcCtx.Im, in)
		case constants.RelationEventGroupMemberRemoved:
			_, err = m.svcCtx.ConversationsModel.SetGroupRemoved(shutdownCtx, in.UserId, in.GroupId, in.Timestamp, in.Version)
		default:
			return nil
		}
		if err == nil {
			return nil
		}
		if status.Code(err) == codes.InvalidArgument {
			zLog.Error("RelationChange.poison", zap.Any("event", in), zap.Error(err))
			return nil
		}
		zLog.Error("RelationChange.reliableState", zap.Any("event", in), zap.Error(err))
		timer := time.NewTimer(delay)
		select {
		case <-shutdownCtx.Done():
			timer.Stop()
			return shutdownCtx.Err()
		case <-timer.C:
		}
		if delay < 30*time.Second {
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}
}

func validGroupRelationEvent(in *mq.RelationChangeTransfer) bool {
	uid, userErr := strconv.ParseUint(in.UserId, 10, 64)
	gid, groupErr := strconv.ParseUint(in.GroupId, 10, 64)
	return userErr == nil && uid > 0 && groupErr == nil && gid > 0 && in.Version > 0 && in.Timestamp > 0
}

type groupConversationEnsurer interface {
	EnsureGroupConversation(context.Context, *im.EnsureGroupConversationReq, ...grpc.CallOption) (*im.EnsureGroupConversationResp, error)
}

func ensureGroupMemberConversation(ctx context.Context, client groupConversationEnsurer, in *mq.RelationChangeTransfer) error {
	_, err := client.EnsureGroupConversation(ctx, &im.EnsureGroupConversationReq{UserId: in.UserId, GroupId: in.GroupId, RelationVersion: in.Version})
	return err
}

// applyCache 按事件类型维护关系缓存。维护失败仅记日志：缓存最终靠 TTL + 读穿透自愈。
func (m *RelationChangeTransfer) applyCache(ctx context.Context, in *mq.RelationChangeTransfer) {
	ApplyRelationEvent(ctx, m.svcCtx.RelationCache, in)
}

// ApplyRelationEvent 把一条关系变更事件应用到关系缓存（版本门/无版本门按事件类型）。
// 消费者与 social 侧 best-effort、以及集成测试共用此派发逻辑。维护失败仅记日志，靠 TTL + 读穿透自愈。
func ApplyRelationEvent(ctx context.Context, rc *relationcache.Cache, in *mq.RelationChangeTransfer) {
	if rc == nil {
		return
	}
	switch in.EventType {
	case constants.RelationEventGroupMemberRemoved:
		if _, err := rc.RemoveGroupMember(ctx, in.GroupId, in.UserId, in.Version); err != nil {
			zLog.Error("RelationChange.RemoveGroupMember", zap.Any("event", in), zap.Error(err))
		}
	case constants.RelationEventGroupDisbanded:
		if err := rc.InvalidateGroup(ctx, in.GroupId); err != nil {
			zLog.Error("RelationChange.InvalidateGroup", zap.Any("event", in), zap.Error(err))
		}
	case constants.RelationEventFriendDeleted:
		// 双向移除：A 的好友集去掉 B，B 的好友集去掉 A。无版本门（删除幂等可交换，跨分区不能用单版本门）
		if _, err := rc.RemoveFriendIfLoaded(ctx, in.FriendA, in.FriendB); err != nil {
			zLog.Error("RelationChange.RemoveFriend A", zap.Any("event", in), zap.Error(err))
		}
		if _, err := rc.RemoveFriendIfLoaded(ctx, in.FriendB, in.FriendA); err != nil {
			zLog.Error("RelationChange.RemoveFriend B", zap.Any("event", in), zap.Error(err))
		}
	case constants.RelationEventFriendAdded:
		// 双向加入：A 的好友集加 B，B 的好友集加 A。无版本门（与删除对称，跨分区不能用单版本门）。
		// 未加载则跳过，留给读穿透从权威 FriendList 重建。
		if _, err := rc.AddFriendIfLoaded(ctx, in.FriendA, in.FriendB); err != nil {
			zLog.Error("RelationChange.AddFriend A", zap.Any("event", in), zap.Error(err))
		}
		if _, err := rc.AddFriendIfLoaded(ctx, in.FriendB, in.FriendA); err != nil {
			zLog.Error("RelationChange.AddFriend B", zap.Any("event", in), zap.Error(err))
		}
	case constants.RelationEventGroupMemberAdded:
		// 群成员新增：同群事件单分区有序 + outbox.id 单调，版本门正确且自带防复活（被踢后旧新增不复活）。
		if _, err := rc.AddGroupMember(ctx, in.GroupId, in.UserId, in.Version); err != nil {
			zLog.Error("RelationChange.AddGroupMember", zap.Any("event", in), zap.Error(err))
		}
	default:
		zLog.Error("RelationChange.unknownType", zap.String("type", in.EventType))
	}
}

// pushRelationChanged 把会话失效帧单推给受影响用户。离线由 ws 侧静默丢弃（上线拉列表已是失效态）。
func (m *RelationChangeTransfer) pushRelationChanged(in *mq.RelationChangeTransfer) {
	switch in.EventType {
	case constants.RelationEventGroupMemberRemoved:
		m.sendRelationFrame(&ws.RelationChanged{
			ReceiverId:     in.UserId,
			EventType:      in.EventType,
			ConversationId: in.GroupId, // 群会话 id 即 groupId
			GroupId:        in.GroupId,
			OperatorId:     in.OperatorId,
			Timestamp:      in.Timestamp,
		})
	case constants.RelationEventGroupMemberAdded:
		m.sendRelationFrame(&ws.RelationChanged{ReceiverId: in.UserId, EventType: in.EventType, ConversationId: in.GroupId, GroupId: in.GroupId, OperatorId: in.OperatorId, Timestamp: in.Timestamp})
	case constants.RelationEventFriendAdded, constants.RelationEventFriendDeleted:
		// 双向通知：各自把对端会话置为失效。单聊会话 id = CombineId(双方)，对称。
		convId := wuid.CombineId(in.FriendA, in.FriendB)
		m.sendRelationFrame(&ws.RelationChanged{
			ReceiverId: in.FriendA, EventType: in.EventType, ConversationId: convId, PeerId: in.FriendB, OperatorId: in.OperatorId, Timestamp: in.Timestamp,
		})
		m.sendRelationFrame(&ws.RelationChanged{
			ReceiverId: in.FriendB, EventType: in.EventType, ConversationId: convId, PeerId: in.FriendA, OperatorId: in.OperatorId, Timestamp: in.Timestamp,
		})
	case constants.RelationEventGroupDisbanded:
		// 解散群无成员名单随事件下发：仅失效缓存，客户端经群列表刷新感知；不在此单推。
	}
}

func (m *RelationChangeTransfer) sendRelationFrame(data *ws.RelationChanged) {
	if data.ReceiverId == "" {
		return
	}
	if err := m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push.relation",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	}); err != nil {
		zLog.Error("RelationChange.send", zap.String("receiver", data.ReceiverId), zap.Error(err))
	}
}
