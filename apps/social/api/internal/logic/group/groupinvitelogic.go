package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GroupInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 邀请群成员
func NewGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLogic {
	return &GroupInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupInviteLogic) GroupInvite(req *types.GroupInviteReq) (resp *types.GroupInviteResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	_, err = l.svcCtx.Social.GroupInvite(l.ctx, &social.GroupInviteReq{
		GroupId:   req.GroupId,
		UserId:    uid,
		FriendIds: req.FriendIds,
	})
	if err != nil {
		zLog.Error("group invite err", zap.Error(err))
		return nil, err
	}
	return
}
