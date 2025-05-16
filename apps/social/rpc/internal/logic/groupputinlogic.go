package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	Group_Members  = "group_members"
	Group_Requests = "group_requests"
)

type GroupPutinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupPutinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutinLogic {
	return &GroupPutinLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupPutin 加入群聊
func (l *GroupPutinLogic) GroupPutin(in *social.GroupPutinReq) (*social.GroupPutinResp, error) {
	//  1. 普通用户申请 ： 如果群无验证直接进入
	//  2. 群成员邀请： 如果群无验证直接进入
	//  3. 群管理员/群创建者邀请：直接进入群

	var (
		inviteGroupMember socialmodels.GroupMembers
		userGroupMember   socialmodels.GroupMembers
		groupInfo         *socialmodels.Groups
		err               error
	)

	//查询用户是否已加入群
	userGroupMember, err = l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.ReqId, in.GroupId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group member by groud id and  req id err %v, req %v, %v", err, in.GroupId, in.ReqId)
	}

	// 如果已经加入
	if userGroupMember.Id != 0 {
		return &social.GroupPutinResp{}, nil
	}

	//查询用户是否已经申请过加入群聊
	groupReq, err := l.svcCtx.GroupRequestsModel.FindByGroupIdAndReqId(l.ctx, in.GroupId, in.ReqId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group req by groud id and user id err %v, req %v, %v", err,
			in.GroupId, in.ReqId)
	}

	// 如果已经申请过
	if groupReq.Id != 0 {
		return &social.GroupPutinResp{}, nil
	}

	// 构建申请
	groupReqTemp := &socialmodels.GroupRequests{
		ReqId:   in.ReqId,   // 申请用户id
		GroupId: in.GroupId, // 请求加入的群id
		ReqMsg: sql.NullString{ // 请求消息
			String: in.ReqMsg,
			Valid:  true,
		},
		ReqTime: sql.NullTime{ //请求时间
			Time:  time.Unix(in.ReqTime, 0),
			Valid: true,
		},
		JoinSource: sql.NullInt64{ //请求来源
			Int64: int64(in.JoinSource),
			Valid: true,
		},
		InviterUserId: sql.NullString{ //要求人，如果是群主和管理员，可以直接进入群聊
			String: in.InviterUid,
			Valid:  true,
		},
		HandleResult: sql.NullInt64{ //处理结果：0未处理
			Int64: int64(constants.NoHandlerResult),
			Valid: true,
		},
	}

	tx := db.GetMysqlConn(db.MYSQL_DB_HICHAT2).Begin()
	defer tx.Commit()

	// 回调处理
	createGroupMember := func() {
		if err != nil {
			return
		}
		err = l.createGroupMember(in, tx)
	}

	//获取群信息
	groupInfo, err = l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find group by groud id err %v, req %v", err, in.GroupId)
	}

	// 不需要验证，直接加入群聊
	if groupInfo.IsVerify == 0 {
		// 不需要，直接通过，加入群聊成员
		defer createGroupMember()

		groupReqTemp.HandleResult = sql.NullInt64{
			Int64: int64(constants.PassHandlerResult),
			Valid: true,
		}

		//创建请求
		return l.createGroupReq(groupReqTemp, true, tx)
	}

	// 主动申请：验证进群方式
	if constants.GroupJoinSource(in.JoinSource) == constants.PutInGroupJoinSource {
		// 申请
		return l.createGroupReq(groupReqTemp, false, tx)
	}

	// 获取邀请人的群成员信息
	inviteGroupMember, err = l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.InviterUid, in.GroupId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "inviter uid not find: %v, groupid: %v", in.InviterUid, in.GroupId)
	}

	//验证是否为管理员或者群主
	if constants.GroupRoleLevel(inviteGroupMember.RoleLevel) == constants.CreatorGroupRoleLevel ||
		constants.GroupRoleLevel(inviteGroupMember.RoleLevel) == constants.ManagerGroupRoleLevel {
		// 是管理者或创建者邀请
		defer createGroupMember()

		groupReqTemp.HandleResult = sql.NullInt64{
			Int64: int64(constants.PassHandlerResult),
			Valid: true,
		}
		groupReqTemp.HandleUserId = sql.NullString{
			String: in.InviterUid,
			Valid:  true,
		}
		return l.createGroupReq(groupReqTemp, true, tx)
	}

	// 其他情况走创建申请，都走审核流程
	return l.createGroupReq(groupReqTemp, false, tx)

}

// createGroupReq 创建加群申请， isPass是否直接加群
func (l *GroupPutinLogic) createGroupReq(groupReq *socialmodels.GroupRequests, isPass bool, tx *gorm.DB) (*social.GroupPutinResp, error) {

	groupReq.HandleTime = time.Now()

	// 申请入库
	res := tx.Table(Group_Requests).Create(&groupReq)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil, errors.New(fmt.Sprintf("groupid: %s create group req failed: err or rows = 0: %s", groupReq.GroupId, res.Error.Error()))
	}

	// 加入群聊，返回群id
	if isPass {
		groupIdInt, err := strconv.Atoi(groupReq.GroupId)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewMsg("群id不合法"), "find group by groud id err %v, req %v", err, groupReq)
		}
		return &social.GroupPutinResp{GroupId: int32(groupIdInt)}, nil
	}

	id, _ := strconv.Atoi(groupReq.GroupId)
	return &social.GroupPutinResp{
		GroupId: int32(id),
	}, nil
}

// createGroupMember 加入群
func (l *GroupPutinLogic) createGroupMember(in *social.GroupPutinReq, tx *gorm.DB) error {
	groupMember := &socialmodels.GroupMembers{
		GroupId:     in.GroupId,
		UserId:      in.ReqId,
		RoleLevel:   int(constants.AtLargeGroupRoleLevel),
		OperatorUid: in.InviterUid,
		JoinTime:    time.Now(),
		JoinSource:  int(in.JoinSource),
		InviterUid:  in.InviterUid,
	}
	res := tx.Table(Group_Members).Create(&groupMember)
	if res.Error != nil || res.RowsAffected == 0 {
		tx.Rollback()
		return errors.Wrapf(xerr.NewDBErr(), "insert friend err %v req %v", res.Error, groupMember)
	}

	return nil
}
