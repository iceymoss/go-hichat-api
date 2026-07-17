package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupInvitationListLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupInvitationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationListLogic {
	return &GroupInvitationListLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupInvitationListLogic) GroupInvitationList(in *social.GroupInvitationListReq) (*social.GroupInvitationListResp, error) {
	if in.ActorUid == "" {
		return nil, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	actor, err := parsePositiveID(in.ActorUid, "actor uid")
	if err != nil {
		return nil, err
	}
	if in.Status < -1 || in.Status > groupInvitationInvalidated {
		return nil, status.Error(codes.InvalidArgument, "invitation status must be between 0 and 4, or -1")
	}
	page, size := in.Page, in.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		return nil, status.Error(codes.InvalidArgument, "page size must not exceed 100")
	}

	query := l.DB.WithContext(l.ctx).Model(&objects.GroupInvitation{}).Where("invitee_uid = ?", actor)
	if in.Status >= 0 {
		query = query.Where("status = ?", in.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to count group invitations")
	}
	var rows []objects.GroupInvitation
	if err := query.Order("created_at DESC, id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&rows).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to list group invitations")
	}
	list := make([]*social.GroupInvitation, len(rows))
	for i := range rows {
		list[i] = invitationProto(&rows[i])
	}
	return &social.GroupInvitationListResp{List: list, Total: total}, nil
}
