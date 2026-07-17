package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
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
	uid := ctxdata.GetUId(l.ctx)
	if id, parseErr := strconv.ParseUint(uid, 10, 64); parseErr != nil || id == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	_, err = l.svcCtx.Social.GroupInvite(l.ctx, &social.GroupInviteReq{
		GroupId:   req.GroupId,
		UserId:    uid,
		FriendIds: req.FriendIds,
		ActorUid:  uid,
	})
	if err != nil {
		return nil, err
	}
	return
}
