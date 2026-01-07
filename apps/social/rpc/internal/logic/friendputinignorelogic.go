package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FriendPutInIgnoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInIgnoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInIgnoreLogic {
	return &FriendPutInIgnoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInIgnore 忽略好友申请（设置status=2）
func (l *FriendPutInIgnoreLogic) FriendPutInIgnore(in *social.FriendPutInIgnoreReq) (*social.FriendPutInIgnoreResp, error) {
	// 获取好友申请记录
	curUidInt, err := strconv.ParseUint(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "parse uid err %v req %v", err, in)
	}

	friendReq, err := l.svcCtx.FriendRequestsModel.FindOne(l.ctx, uint64(in.FriendReqId), in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend request err %v req %v", err, in)
	}

	// 检查是否是当前用户的好友申请（我收到的申请：req_uid = curUid，我发起的申请：user_id = curUid）
	if friendReq.ReqUid != curUidInt && friendReq.UserId != curUidInt {
		return nil, errors.Wrapf(xerr.NewDBErr(), "friend request not belong to current user")
	}

	// 设置status=2（忽略不显示）
	friendReq.Status = 2

	// 使用事务更新记录
	err = l.svcCtx.FriendRequestsModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		return l.svcCtx.FriendRequestsModel.Update(ctx, session, friendReq)
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update friend request status err %v req %v", err, in)
	}

	return &social.FriendPutInIgnoreResp{}, nil
}
