package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindGroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindGroupListLogic {
	return &FindGroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindGroupListLogic) FindGroupList(in *social.FindGroupListReq) (*social.FindGroupListResp, error) {
	groupList, err := l.svcCtx.GroupsModel.ListByGroupIds(l.ctx, in.Ids)
	if err != nil {
		return nil, err
	}

	list := make([]*social.Groups, 0, len(groupList))
	for _, v := range groupList {
		list = append(list, &social.Groups{
			Id:         strconv.Itoa(v.Id),
			Name:       v.Name,
			Icon:       v.Icon,
			Status:     int32(v.Status),
			CreatorUid: v.CreatorUid,
		})
	}

	return &social.FindGroupListResp{
		List: list,
	}, nil
}
