package group

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGroupListLogic 用户群列表
func NewGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupListLogic {
	return &GroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupListLogic) GroupList(req *types.GroupListRep) (resp *types.GroupListResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.Social.GroupList(l.ctx, &social.GroupListReq{UserId: uid})
	if err != nil {
		return nil, err
	}

	list := make([]*types.Groups, 0, len(res.List))
	for _, v := range res.List {
		list = append(list, &types.Groups{
			Id:              v.Id,
			Name:            v.Name,
			Icon:            v.Icon,
			Status:          int64(v.Status),
			GroupType:       int64(v.GroupType),
			CreateUid:       v.CreatorUid,
			IsVerify:        v.IsVerify,
			Notification:    v.Notification,
			NotificationUid: v.NotificationUid,
		})
	}

	resp = &types.GroupListResp{List: list}

	return
}
