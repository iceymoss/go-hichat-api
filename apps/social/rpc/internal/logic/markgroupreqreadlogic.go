package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type MarkGroupReqReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkGroupReqReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkGroupReqReadLogic {
	return &MarkGroupReqReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkGroupReqRead 把"我（群主/管理员）管理的群"收到的入群申请全部标记已读（receiver_read=1）。
// 与好友申请进入列表即全部标已读的语义一致：badge=未读新申请，查看后清零。
func (l *MarkGroupReqReadLogic) MarkGroupReqRead(in *social.MarkGroupReqReadReq) (*social.MarkGroupReqReadResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	if len(in.RequestIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "group request ids are required")
	}
	var counts receiptCounts
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := markReceiptsRead(tx, actor, receiptTypeGroup, in.RequestIds); err != nil {
			return err
		}
		var err error
		counts, err = countUnreadReceipts(tx, actor, true)
		return err
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark group request receipts read")
	}
	return &social.MarkGroupReqReadResp{Count: counts.Total, Apply: counts.Apply, Result: counts.Result, Invite: counts.Invite}, nil
}
