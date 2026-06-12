package logic

import (
	"context"
	"strings"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	socialPb "github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	userPb "github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/wuid"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

const Identify = "hichat2.com"

type GetConversationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationsLogic) GetConversations(req *types.GetConversationsReq) (resp *types.GetConversationsResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.IM.GetConversations(l.ctx, &im.GetConversationsReq{
		UserId: uid,
	})
	if err != nil {
		zLog.Error("get conversations failed", zap.Error(err))
		return nil, err
	}

	// 1. 构建会话列表，收集需要查询的用户ID和群ID
	peerIds := make([]string, 0)  // 私聊对方 userId
	groupIds := make([]string, 0) // 群聊 groupId
	conversations := make(map[string]types.Conversation, len(res.ConversationList))

	for k, v := range res.ConversationList {
		if !v.IsShow {
			continue
		}
		var message types.ChatLog
		if v.Msg != nil {
			message = types.ChatLog{
				ConversationId: v.Msg.ConversationId,
				SendId:         v.Msg.SendId,
				RecvId:         v.Msg.RecvId,
				MsgType:        v.Msg.MsgType,
				MsgContent:     v.Msg.MsgContent,
				ChatType:       v.ChatType,
				SendTime:       v.Msg.SendTime,
			}
		}

		conv := types.Conversation{
			ConversationId: v.ConversationId,
			ChatType:       v.ChatType,
			IsShow:         v.IsShow,
			IsTop:          v.IsTop,
			IsMute:         v.IsMute,
			HasAtMe:        v.HasAtMe,
			Seq:            v.Seq,
			Read:           v.ToRead,
			Msg:            message,
		}

		if v.ChatType == 2 { // 群聊
			groupIds = append(groupIds, v.ConversationId)
		} else { // 私聊
			peerId := extractPeerId(v.ConversationId, uid)
			if peerId != "" {
				peerIds = append(peerIds, peerId)
			}
		}

		conversations[k] = conv
	}

	// 2. 批量查询私聊对方用户信息
	if len(peerIds) > 0 {
		userResp, err := l.svcCtx.User.FindUser(l.ctx, &userPb.FindUserReq{
			Ids: peerIds,
		})
		if err == nil && userResp != nil {
			userMap := make(map[string]*userPb.UserEntity)
			for _, u := range userResp.User {
				userMap[u.Id] = u
			}
			// 回填到会话
			for k, conv := range conversations {
				if conv.ChatType != 1 {
					continue
				}
				peerId := extractPeerId(conv.ConversationId, uid)
				if u, ok := userMap[peerId]; ok {
					conv.TargetName = u.Nickname
					conv.TargetAvatar = u.Avatar
					conversations[k] = conv
				}
			}
		}

		// 补充好友备注名
		friendResp, err := l.svcCtx.Social.FriendList(l.ctx, &socialPb.FriendListReq{
			UserId: uid,
		})
		if err == nil && friendResp != nil {
			remarkMap := make(map[string]string)
			for _, f := range friendResp.List {
				if f.Remark != "" {
					remarkMap[f.FriendUid] = f.Remark
				}
			}
			for k, conv := range conversations {
				if conv.ChatType != 1 {
					continue
				}
				peerId := extractPeerId(conv.ConversationId, uid)
				if remark, ok := remarkMap[peerId]; ok {
					conv.TargetName = remark
					conversations[k] = conv
				}
			}
		}
	}

	// 3. 批量查询群信息
	if len(groupIds) > 0 {
		groupMap := make(map[string]*socialPb.Groups)
		if groupResp, err := l.svcCtx.Social.GroupList(l.ctx, &socialPb.GroupListReq{
			UserId: uid,
		}); err == nil && groupResp != nil {
			for _, g := range groupResp.List {
				groupMap[g.Id] = g
			}
		}
		for k, conv := range conversations {
			if conv.ChatType != 2 {
				continue
			}
			if g, ok := groupMap[conv.ConversationId]; ok {
				conv.TargetName = g.Name
				conv.TargetAvatar = g.Icon
				conversations[k] = conv
				continue
			}
			// 已退群/被移出的群不在 GroupList 里：按 groupId 直查群信息补名字/头像，
			// 否则前端会话只能显示群号（如 "15"）。GroupDetail 按 id 查、不校验成员身份。
			if detail, derr := l.svcCtx.Social.GroupDetail(l.ctx, &socialPb.GroupDetailReq{
				GroupId: conv.ConversationId,
			}); derr == nil && detail != nil && detail.Group != nil {
				conv.TargetName = detail.Group.Name
				conv.TargetAvatar = detail.Group.Icon
				conversations[k] = conv
			}
		}
	}

	return &types.GetConversationsResp{ConversationList: conversations}, nil
}

// extractPeerId 从会话ID中提取对方用户ID
func extractPeerId(conversationId, myId string) string {
	// 优先用 wuid.CombineId 的格式：小ID_大ID
	parts := strings.Split(conversationId, "_")
	if len(parts) == 2 {
		if parts[0] == myId {
			return parts[1]
		}
		return parts[0]
	}
	return ""
}

// 确保 wuid 包被导入（用于未来可能的验证）
var _ = wuid.CombineId
