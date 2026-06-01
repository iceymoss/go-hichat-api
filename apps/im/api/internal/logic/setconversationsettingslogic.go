package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type SetConversationSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 设置会话置顶/免打扰
func NewSetConversationSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConversationSettingsLogic {
	return &SetConversationSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetConversationSettingsLogic) SetConversationSettings(req *types.SetConversationSettingsReq) (resp *types.SetConversationSettingsResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	_, err = l.svcCtx.IM.SetConversationSettings(l.ctx, &im.SetConversationSettingsReq{
		UserId:         uid,
		ConversationId: req.ConversationId,
		IsTop:          req.IsTop,
		IsMute:         req.IsMute,
	})
	if err != nil {
		zLog.Error("set conversation settings failed", zap.Error(err))
		return nil, err
	}

	return &types.SetConversationSettingsResp{}, nil
}
