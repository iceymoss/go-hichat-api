package logic

import (
	"context"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
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

	//todo: 事务无效，待处理
	err = l.svcCtx.GroupsModel.Transact(l.ctx, func(ctx context.Context, db *gorm.DB) error {
		groupID, insertErr := l.svcCtx.GroupsModel.Insert(l.ctx, groups)
		if insertErr != nil {
			return errors.Wrapf(xerr.NewDBErr(), "insert group err %v req %v", err, in)
		}

		fmt.Println("id:", groupID)

		//将群主加入群里
		_, err = l.svcCtx.GroupMembersModel.Insert(l.ctx, &socialmodels.GroupMembers{
			GroupId:     strconv.Itoa(groupID),
			UserId:      strconv.Itoa(creatorUidInt),
			RoleLevel:   int(constants.CreatorGroupRoleLevel),
			JoinTime:    time.Now(),
			JoinSource:  0,
			InviterUid:  strconv.Itoa(creatorUidInt),
			OperatorUid: strconv.Itoa(creatorUidInt),
		})
		if err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "insert group member err %v req %v", err, in)
		}
		return nil
	})

	return &social.GroupCreateResp{}, err
}
