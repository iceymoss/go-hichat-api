package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type FriendDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendDeleteLogic {
	return &FriendDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendDeleteLogic) FriendDelete(in *social.FriendDeleteReq) (*social.FriendDeleteResp, error) {
	var idsToDelete []uint64

	relation, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.FriendUid)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend relation err:%v req:%v", err, in)
	}
	if relation != nil {
		idsToDelete = append(idsToDelete, relation.Id)
	}

	reverse, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.FriendUid, in.UserId)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find reverse friend relation err:%v req:%v", err, in)
	}
	if reverse != nil {
		idsToDelete = append(idsToDelete, reverse.Id)
	}

	if len(idsToDelete) == 0 {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "friend relation not found uid:%s fid:%s", in.UserId, in.FriendUid)
	}

	// 双向删除 + 同事务写 outbox（删好友事件），提交后 best-effort 同步关系缓存
	err = emitRelationChange(l.ctx, l.svcCtx, constants.RelationEventFriendDeleted, "",
		&mq.RelationChangeTransfer{FriendA: in.UserId, FriendB: in.FriendUid, OperatorId: in.UserId},
		func(tx *gorm.DB) error {
			return tx.Table("friends").Where("id IN ?", idsToDelete).Delete(nil).Error
		})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "delete friend relation err:%v req:%v", err, in)
	}

	return &social.FriendDeleteResp{}, nil
}
