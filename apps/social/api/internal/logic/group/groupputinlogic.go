package group

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupPutInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInLogic {
	return &GroupPutInLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GroupPutInLogic) GroupPutIn(req *types.GroupPutInReq) (*types.GroupPutInResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if id, err := strconv.ParseUint(uid, 10, 64); err != nil || id == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	if req.Token != "" {
		res, err := l.svcCtx.Social.GroupJoinByToken(l.ctx, &social.GroupJoinByTokenReq{UserId: uid, Token: req.Token, ReqMsg: req.ReqMsg})
		if err != nil {
			return nil, err
		}
		return &types.GroupPutInResp{GroupId: int(res.GroupId), GroupIdString: strconv.FormatInt(int64(res.GroupId), 10), IsPass: int(res.IsPass), Status: res.IsPass}, nil
	}

	// Legacy identity/source fields are intentionally ignored for public direct applications.
	res, err := l.svcCtx.Social.GroupPutin(l.ctx, &social.GroupPutinReq{
		GroupId: req.GroupId, ReqId: uid, ReqMsg: req.ReqMsg, ReqTime: time.Now().Unix(),
		JoinSource: 1, ActorUid: uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupPutInResp{
		GroupId: int(res.GroupId), GroupIdString: res.GroupIdString, IsPass: int(res.IsPass), RequestId: res.RequestId,
		Status: res.Status, AlreadyPending: res.AlreadyPending, AlreadyMember: res.AlreadyMember,
	}, nil
}
