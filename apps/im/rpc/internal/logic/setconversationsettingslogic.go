package logic

import (
	"context"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type SetConversationSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetConversationSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConversationSettingsLogic {
	return &SetConversationSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetConversationSettings 设置会话置顶/免打扰（每用户每会话），全量覆盖 IsTop/IsMute
func (l *SetConversationSettingsLogic) SetConversationSettings(in *im.SetConversationSettingsReq) (*im.SetConversationSettingsResp, error) {
	// 获取用户会话
	conversations, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find conversation by userId err %v req %v", err, in.UserId)
	}

	if conversations.ConversationList == nil {
		conversations.ConversationList = make(map[string]*models.Conversation)
	}

	if conv, ok := conversations.ConversationList[in.ConversationId]; ok && conv != nil {
		// 已存在：保留其余字段，仅更新置顶/免打扰
		conv.IsTop = in.IsTop
		conv.IsMute = in.IsMute
	} else {
		// 不存在：建最小条目，保证设置可落库
		conversations.ConversationList[in.ConversationId] = &models.Conversation{
			ConversationId: in.ConversationId,
			IsShow:         true,
			IsTop:          in.IsTop,
			IsMute:         in.IsMute,
		}
	}

	if _, err = l.svcCtx.ConversationsModel.Update(l.ctx, conversations); err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update conversation settings err %v req %v", err, in)
	}

	return &im.SetConversationSettingsResp{}, nil
}
