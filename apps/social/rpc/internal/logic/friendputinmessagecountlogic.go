package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FriendPutInMessageCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInMessageCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInMessageCountLogic {
	return &FriendPutInMessageCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInMessageCount 获取好友申请消息数量
// 规则：
// 1. 我发起的申请：handle_result=2（已拒绝）且 status=1（正常显示）的数量
// 2. 我收到的申请：status=1（正常显示）且 handle_result=0（待处理）的数量
func (l *FriendPutInMessageCountLogic) FriendPutInMessageCount(in *social.FriendPutInMessageCountReq) (*social.FriendPutInMessageCountResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	counts, err := countUnreadReceipts(l.svcCtx.DB.WithContext(l.ctx), actor, false)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to count friend request receipts")
	}

	return &social.FriendPutInMessageCountResp{
		Count: counts.Total, Apply: counts.Apply, Result: counts.Result,
	}, nil
}
