package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrGroupReqBeforePass   = xerr.NewMsg("请求已通过")
	ErrGroupReqBeforeRefuse = xerr.NewMsg("请求已拒绝")
)

type GroupPutInHandleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInHandleLogic {
	return &GroupPutInHandleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupPutInHandle 处理加群申请
func (l *GroupPutInHandleLogic) GroupPutInHandle(in *social.GroupPutInHandleReq) (*social.GroupPutInHandleResp, error) {
	groupReq, err := l.svcCtx.GroupRequestsModel.FindOne(l.ctx, int64(in.GroupReqId))
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend req err %v req %v", err, in.GroupReqId)
	}

	// 权限校验：只有群主/管理员可以审批入群申请
	handleMember, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.HandleUid, groupReq.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "handle user not in group")
	}
	if handleMember.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only admin/owner can handle group requests")
	}

	switch constants.HandlerResult(groupReq.HandleResult.Int64) {
	case constants.PassHandlerResult:
		return nil, errors.WithStack(ErrGroupReqBeforePass)
	case constants.RefuseHandlerResult:
		return nil, errors.WithStack(ErrGroupReqBeforeRefuse)
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()

	groupReq.HandleTime = time.Now()
	groupReq.HandleResult = sql.NullInt64{
		Int64: int64(in.HandleResult),
		Valid: true,
	}
	groupReq.HandleUserId = sql.NullString{
		String: in.HandleUid,
		Valid:  true,
	}

	groupReq.HandleTime = time.Now()

	//更新申请状态
	res := tx.Table(constants.GroupRequests).Where("id = ?", groupReq.Id).Save(&groupReq)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("update err groupRedId:" + groupReq.ReqId)
	}

	if constants.HandlerResult(groupReq.HandleResult.Int64) != constants.PassHandlerResult {
		//拒绝加入群
		tx.Commit()
		return &social.GroupPutInHandleResp{GroupId: groupReq.GroupId, ReqId: groupReq.ReqId}, nil
	}

	//插入群成员表
	groupMember := &socialmodels.GroupMembers{
		GroupId:     groupReq.GroupId,
		UserId:      groupReq.ReqId,
		RoleLevel:   int(constants.AtLargeGroupRoleLevel),
		OperatorUid: in.HandleUid,
		JoinTime:    time.Now(),
		InviterUid:  groupReq.InviterUserId.String,
		JoinSource:  int(groupReq.JoinSource.Int64),
	}

	res = tx.Create(&groupMember)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("join  group err:" + groupReq.ReqId + " groupId" + groupReq.GroupId)
	}

	// 同事务写 outbox（群成员新增事件），relay 投递后消费端版本门把新成员加入群成员集
	if err := emitRelationChangeInTx(tx, l.svcCtx, constants.RelationEventGroupMemberAdded, groupReq.GroupId,
		&mq.RelationChangeTransfer{GroupId: groupReq.GroupId, UserId: groupReq.ReqId, OperatorId: in.HandleUid}); err != nil {
		tx.Rollback()
		return nil, err
	}

	//为新成员添加会话
	//_, err = l.svcCtx.IM.CreateGroupConversation(l.ctx, &im.CreateGroupConversationReq{
	//	GroupId:  groupReq.GroupId,
	//	CreateId: groupReq.ReqId,
	//})
	//if err != nil {
	//	zLog.Error("GroupCreate.CreateGroupConversation: create group conversation failed", zap.Any("uid", groupReq.ReqId), zap.Error(err))
	//	tx.Rollback()
	//	return nil, err
	//}

	tx.Commit()

	return &social.GroupPutInHandleResp{
		GroupId: groupReq.GroupId,
		ReqId:   groupReq.ReqId,
	}, err
}
