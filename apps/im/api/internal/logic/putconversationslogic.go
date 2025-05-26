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

type PutConversationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新会话
func NewPutConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PutConversationsLogic {
	return &PutConversationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PutConversationsLogic) PutConversations(req *types.PutConversationsReq) (resp *types.PutConversationsResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	conversationList := make(map[string]*im.Conversation, len(req.ConversationList))
	for k, v := range req.ConversationList {
		conversationList[k] = &im.Conversation{
			ConversationId: v.ConversationId,
			ChatType:       v.ChatType,
			TargetId:       "",
			IsShow:         v.IsShow,
			Seq:            v.Seq,
			ToRead:         v.Read,
			Read:           v.Read,
		}
	}
	_, err = l.svcCtx.IM.PutConversations(l.ctx, &im.PutConversationsReq{
		// todo:id
		Id:               "",
		UserId:           uid,
		ConversationList: conversationList,
	})
	if err != nil {
		zLog.Error("put conversations failed", zap.Error(err))
		return nil, err
	}

	return
}
