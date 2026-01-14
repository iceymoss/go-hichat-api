package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupTransferOwnerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 转让群主（仅群主）
func NewGroupTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupTransferOwnerLogic {
	return &GroupTransferOwnerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupTransferOwnerLogic) GroupTransferOwner(req *types.GroupTransferOwnerReq) (resp *types.GroupTransferOwnerResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	// KeepOldOwnerAsAdmin 是 bool 类型，默认 true（服务端兜底）
	// 由于是 bool 类型，无法区分"未传"和"传 false"
	// 约定：默认 true（保留原群主为管理员）
	// 如果前端需要设置为 false，需要显式传 false
	keepOld := req.KeepOldOwnerAsAdmin
	// 如果为 false，可能是显式传的 false，也可能是未传（默认 false）
	// 为了安全，我们默认 true（保留原群主为管理员）
	// 如果前端需要设置为 false，必须显式传 false
	// 这里直接使用 req.KeepOldOwnerAsAdmin，让 RPC 层处理默认值

	_, err = l.svcCtx.Social.GroupTransferOwner(l.ctx, &social.GroupTransferOwnerReq{
		UserId:              uid,
		GroupId:             req.GroupId,
		NewOwnerId:          req.NewOwnerId,
		KeepOldOwnerAsAdmin: keepOld,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupTransferOwnerResp{}, nil
}
