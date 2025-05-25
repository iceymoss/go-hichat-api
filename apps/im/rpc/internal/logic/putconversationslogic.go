package logic

import (
	"context"
	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PutConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPutConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PutConversationsLogic {
	return &PutConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PutConversations 更新用户会话
func (l *PutConversationsLogic) PutConversations(in *im.PutConversationsReq) (*im.PutConversationsResp, error) {
	// 获取用户会话
	conversations, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find conversation by userId err %v req %v", err, in.UserId)
	}

	// 会话为空，则生成一个会话
	if conversations.ConversationList == nil {
		conversations.ConversationList = make(map[string]*models.Conversation)
	}

	for k, i := range in.ConversationList {
		var oldTotal int
		if conversations.ConversationList[k] != nil {
			oldTotal = conversations.ConversationList[k].Total
		}

		conversations.ConversationList[k] = &models.Conversation{
			ConversationId: i.ConversationId,
			ChatType:       constants.ChatType(i.ChatType),
			IsShow:         i.IsShow,
			// 更新最新的已读总记录
			//	i.Read前端传入某一个会话的消息已经被接收站读取的数量
			Total: int(i.Read) + oldTotal,
			Seq:   i.Seq,
		}
	}

	_, err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update conversation err %v req %v", err, conversations)
	}

	return &im.PutConversationsResp{}, err
}
