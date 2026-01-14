package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUpdateLogic {
	return &GroupUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupUpdateLogic) GroupUpdate(req *types.GroupUpdateReq) (resp *types.GroupUpdateResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	// IsVerify 是 int32 类型，-1 表示不修改，0=不需要验证，1=需要验证
	// RPC 层逻辑：只有 IsVerify == 0 或 1 时才会修改，其他值（包括 -1）不修改
	// 由于是 int32 类型，无法区分"未传"和"传 0"
	// 约定：如果 IsVerify == 0 且没有其他字段要更新，认为未传，传 -1（不修改）
	// 如果 IsVerify == 0 且有其他字段要更新，传 0（不需要验证）
	// 如果 IsVerify == 1，传 1（需要验证）
	// 如果 IsVerify == -1，传 -1（不修改）
	isVerify := req.IsVerify
	if isVerify == 0 && req.Name == "" && req.Icon == "" && req.Notification == "" {
		// IsVerify == 0 且没有其他字段，认为未传，使用 -1（不修改）
		isVerify = -1
	}

	_, err = l.svcCtx.Social.GroupUpdate(l.ctx, &social.GroupUpdateReq{
		UserId:       uid,
		GroupId:      req.GroupId,
		Name:         req.Name,
		Icon:         req.Icon,
		Notification: req.Notification,
		IsVerify:     isVerify,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupUpdateResp{}, nil
}
