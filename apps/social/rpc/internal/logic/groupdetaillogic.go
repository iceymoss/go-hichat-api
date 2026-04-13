package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupDetailLogic {
	return &GroupDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupDetailLogic) GroupDetail(in *social.GroupDetailReq) (*social.GroupDetailResp, error) {
	// 1. Get Group
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found %v", in.GroupId)
	}

	// 2. Get Members
	members, err := l.svcCtx.GroupMembersModel.ListByGroupId(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list members err %v", err)
	}

	// 记录rpc服务调用rpc服务
	//l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
	//	Name:  "",
	//	Phone: "",
	//	Ids:   nil,
	//	Email: "",
	//})

	// 3. Convert to proto types
	groupProto := &social.Groups{
		Id:              strconv.Itoa(group.Id),
		Name:            group.Name,
		Icon:            group.Icon,
		Status:          int32(group.Status),
		CreatorUid:      group.CreatorUid,
		GroupType:       int32(group.GroupType),
		IsVerify:        group.IsVerify == 1, // Model IsVerify is int, proto is bool
		Notification:    group.Notification,
		NotificationUid: strconv.Itoa(group.NotificationUid),
	}

	var membersProto []*social.GroupMembers
	for _, m := range members {
		membersProto = append(membersProto, &social.GroupMembers{
			Id:            int32(m.Id),
			GroupId:       m.GroupId,
			UserId:        m.UserId,
			RoleLevel:     int32(m.RoleLevel),
			JoinTime:      m.JoinTime.Unix(),
			JoinSource:    int32(m.JoinSource),
			InviterUid:    m.InviterUid,
			OperatorUid:   m.OperatorUid,
			GroupNickname: m.GroupNickname,
			GroupRemark:   m.GroupRemark,
		})
	}

	return &social.GroupDetailResp{
		Group:   groupProto,
		Members: membersProto,
	}, nil
}
