package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupRequestMessageCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupRequestMessageCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupRequestMessageCountLogic {
	return &GroupRequestMessageCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupRequestMessageCountLogic) GroupRequestMessageCount(in *social.GroupRequestMessageCountReq) (*social.GroupRequestMessageCountResp, error) {
	counts, err := countUnreadReceipts(l.svcCtx.DB.WithContext(l.ctx), in.UserId, true)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to count group request receipts")
	}
	return &social.GroupRequestMessageCountResp{Count: counts.Total, Apply: counts.Apply, Result: counts.Result, Invite: counts.Invite}, nil
}
