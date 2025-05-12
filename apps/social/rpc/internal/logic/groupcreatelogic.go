package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GroupCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGroupCreateLogic 群业务：创建群，修改群，群公告，申请群，用户群列表，群成员，申请群，群退出
func NewGroupCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupCreateLogic {
	return &GroupCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupCreate 创建群聊
func (l *GroupCreateLogic) GroupCreate(in *social.GroupCreateReq) (*social.GroupCreateResp, error) {
	creatorUidInt, err := strconv.Atoi(in.CreatorUid)
	if err != nil {
		errors.Wrapf(xerr.NewMsg("创建失败"), "cur creatorUid type err %v req %v", err, in)
	}

	groups := &socialmodels.Groups{
		Name:            in.Name,
		Icon:            in.Icon,
		CreatorUid:      in.CreatorUid,
		IsVerify:        1,
		Status:          0,
		GroupType:       1,
		Notification:    "",
		NotificationUid: 0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()
	res := tx.Table("groups").Create(&groups)
	if res.Error != nil || res.RowsAffected == 0 {
		tx.Rollback()
		zLog.Error("create group err", zap.Any("err", res.Error))
		return nil, res.Error
	}

	groupMember := &socialmodels.GroupMembers{
		GroupId:     strconv.Itoa(groups.Id),
		UserId:      strconv.Itoa(creatorUidInt),
		RoleLevel:   int(constants.CreatorGroupRoleLevel),
		JoinTime:    time.Now(),
		JoinSource:  0,
		InviterUid:  strconv.Itoa(creatorUidInt),
		OperatorUid: strconv.Itoa(creatorUidInt),
	}
	res = tx.Table("group_members").Create(&groupMember)
	if res.Error != nil || res.RowsAffected == 0 {
		tx.Rollback()
		zLog.Error("insert group err", zap.Any("err", res.Error))
		return nil, res.Error
	}

	//提交事务
	tx.Commit()

	return &social.GroupCreateResp{}, err
}
