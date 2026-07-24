package logic

import (
	"context"
	"database/sql"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUsersLogic {
	return &GroupUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupUsers 获取群成员
func (l *GroupUsersLogic) GroupUsers(in *social.GroupUsersReq) (*social.GroupUsersResp, error) {
	groupMembers, err := l.svcCtx.GroupMembersModel.ListByGroupId(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list group member err %v req %v", err, in.GroupId)
	}

	respList := make([]*social.GroupMembers, 0, len(groupMembers))
	for _, v := range groupMembers {
		respList = append(respList, &social.GroupMembers{
			Id:          int32(v.Id),
			GroupId:     v.GroupId,
			UserId:      v.UserId,
			RoleLevel:   int32(v.RoleLevel),
			JoinTime:    nullableGroupMemberJoinTime(v.JoinTime),
			JoinSource:  int32(v.JoinSource.Int64),
			InviterUid:  v.InviterUid.String,
			OperatorUid: v.OperatorUid.String,
		})
	}
	return &social.GroupUsersResp{
		List: respList,
	}, nil
}

func nullableGroupMemberJoinTime(value sql.NullTime) int64 {
	if !value.Valid {
		return 0
	}
	return value.Time.Unix()
}
