package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupListLogic {
	return &GroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupList 用户群列表
func (l *GroupListLogic) GroupList(in *social.GroupListReq) (*social.GroupListResp, error) {
	userGroup, err := l.svcCtx.GroupMembersModel.ListByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list group member err %v req %v", err, in.UserId)
	}
	if len(userGroup) == 0 {
		return &social.GroupListResp{}, nil
	}

	ids := make([]string, 0, len(userGroup))
	for _, v := range userGroup {
		ids = append(ids, v.GroupId)
	}

	groups, err := l.svcCtx.GroupsModel.ListByGroupIds(l.ctx, ids)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list group err %v req %v", err, ids)
	}

	respList := make([]*social.Groups, 0, len(userGroup))
	for _, v := range groups {
		var IsVerify bool
		if v.IsVerify > 0 {
			IsVerify = true
		}
		respList = append(respList, &social.Groups{
			Id:              string(v.Id),
			Name:            v.Name,
			Icon:            v.Icon,
			Status:          int32(v.Status.Int64),
			CreatorUid:      string(v.CreatorUid),
			GroupType:       int32(v.GroupType),
			IsVerify:        IsVerify,
			Notification:    v.Notification.String,
			NotificationUid: string(v.NotificationUid.Int64),
		})
	}

	return &social.GroupListResp{
		List: respList,
	}, nil
}
