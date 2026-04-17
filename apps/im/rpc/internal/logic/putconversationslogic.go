package logic

import (
	"context"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
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

// PutConversations 更新用户会话，主要用于更新未读消息数量
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

	// 当删除会话(isShow=false)时，需要获取该会话的实际消息数，
	// 这样 GET 时 total == conversation.Total，不会因为"有未读消息"而被强制恢复显示
	convIds := make([]string, 0)
	for k := range in.ConversationList {
		convIds = append(convIds, k)
	}
	convTotals := make(map[string]int)
	if len(convIds) > 0 {
		convList, _ := l.svcCtx.ConversationModel.ListByConversationIds(l.ctx, convIds)
		for _, c := range convList {
			convTotals[c.ConversationId] = int(c.Total)
		}
	}

	for k, i := range in.ConversationList {
		var oldTotal int
		if conversations.ConversationList[k] != nil {
			oldTotal = conversations.ConversationList[k].Total
		}

		newTotal := int(i.Read) + oldTotal
		// 删除会话时，将 Total 设为实际消息数，防止 GET 时误判为有未读消息
		if !i.IsShow {
			if actualTotal, ok := convTotals[k]; ok && actualTotal > newTotal {
				newTotal = actualTotal
			}
		}

		conversations.ConversationList[k] = &models.Conversation{
			ConversationId: i.ConversationId,
			ChatType:       constants.ChatType(i.ChatType),
			IsShow:         i.IsShow,
			Total:          newTotal,
			Seq:            i.Seq,
		}
	}

	_, err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update conversation err %v req %v", err, conversations)
	}

	return &im.PutConversationsResp{}, err
}
