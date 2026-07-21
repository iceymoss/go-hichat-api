package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkGroupReqReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkGroupReqReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkGroupReqReadLogic {
	return &MarkGroupReqReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkGroupReqRead 把"我（群主/管理员）管理的群"收到的入群申请全部标记已读（receiver_read=1）。
// 与好友申请进入列表即全部标已读的语义一致：badge=未读新申请，查看后清零。
func (l *MarkGroupReqReadLogic) MarkGroupReqRead(in *social.MarkGroupReqReadReq) (*social.MarkGroupReqReadResp, error) {
	if in.UserId == "" {
		return &social.MarkGroupReqReadResp{}, nil
	}

	conn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)

	// 我管理的群（带 role_level，ListByUserId 不含 role_level，直接查）
	var managed []string
	if err := conn.WithContext(l.ctx).Table("group_members").
		Where("user_id = ? AND role_level >= ?", in.UserId, int(constants.ManagerGroupRoleLevel)).
		Pluck("group_id", &managed).Error; err != nil {
		return nil, err
	}
	if len(managed) == 0 {
		return &social.MarkGroupReqReadResp{}, nil
	}

	if err := conn.WithContext(l.ctx).Table("group_requests").
		Where("group_id in ? AND receiver_read = ?", managed, 0).
		Update("receiver_read", 1).Error; err != nil {
		return nil, err
	}

	return &social.MarkGroupReqReadResp{}, nil
}
