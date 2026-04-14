package logic

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

const Identify = "hichat2.com"

type GetConversationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetConversationsLogic 获取会话
func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetConversations 获取用户会话(聊天列表)列表
func (l *GetConversationsLogic) GetConversations(req *types.GetConversationsReq) (resp *types.GetConversationsResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.IM.GetConversations(l.ctx, &im.GetConversationsReq{
		UserId: uid,
	})
	if err != nil {
		zLog.Error("get conversations failed", zap.Error(err))
		return nil, err
	}

	// 会话列表需要展示最后一条聊天信息,需要从会话详情中获取会话last_msg
	// 过滤掉 isShow=false 的已删除会话
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

		conversations[k] = types.Conversation{
			ConversationId: v.ConversationId,
			ChatType:       v.ChatType,
			IsShow:         v.IsShow,
			Seq:            v.Seq,
			Read:           v.ToRead, //未读数量
			Msg:            message,
		}
		fmt.Printf("conversations: %+v\n", conversations[k])
	}

	resp = &types.GetConversationsResp{ConversationList: conversations}

	return
}
