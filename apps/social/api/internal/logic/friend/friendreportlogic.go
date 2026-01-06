package friend

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendReportLogic {
	return &FriendReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 举报好友，记录到 friend_reports
func (l *FriendReportLogic) FriendReport(req *types.FriendReportReq) (*types.FriendReportResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}

	_, err := l.svcCtx.Social.FriendReport(l.ctx, &socialclient.FriendReportReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		Reason:    req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendReportResp{}, nil
}
