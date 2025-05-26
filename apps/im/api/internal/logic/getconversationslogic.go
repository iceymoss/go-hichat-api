package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetConversationsLogic) GetConversations(req *types.GetConversationsReq) (resp *types.GetConversationsResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.IM.GetConversations(l.ctx, &im.GetConversationsReq{
		UserId: uid,
	})
	if err != nil {
		zLog.Error("get conversations failed", zap.Error(err))
		return nil, err
	}

	conversations := make(map[string]types.Conversation, len(res.ConversationList))
	for k, v := range res.ConversationList {
		conversations[k] = types.Conversation{
			ConversationId: v.ConversationId,
			ChatType:       v.ChatType,
			IsShow:         v.IsShow,
			Seq:            v.Seq,
			Read:           v.Read,
		}
	}

	resp = &types.GetConversationsResp{ConversationList: conversations}

	return
}
