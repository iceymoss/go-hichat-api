package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 申请进群列表
func NewGroupPutInListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInListLogic {
	return &GroupPutInListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupPutInListLogic) GroupPutInList(req *types.GroupPutInListRep) (resp *types.GroupPutInListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
