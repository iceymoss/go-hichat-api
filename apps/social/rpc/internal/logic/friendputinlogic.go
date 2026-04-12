package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/utils"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FriendPutInLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInLogic {
	return &FriendPutInLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutIn 好友业务：请求好友
func (l *FriendPutInLogic) FriendPutIn(in *social.FriendPutInReq) (*social.FriendPutInResp, error) {
	// 申请人是否与目标是好友关系
	friends, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.ReqUid)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friends by uid and fid err %v req %v ", err, in)
	}
	if friends != nil {
		return &social.FriendPutInResp{}, err
	}

	// 是否已经有过申请，申请是不成功，没有完成
	friendReqs, err := l.svcCtx.FriendRequestsModel.FindByReqUidAndUserId(l.ctx, in.ReqUid, in.UserId)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friendsRequest by rid and uid err %v req %v ", err, in)
	}

	uidInt, err := strconv.Atoi(in.UserId)
	if err != nil {
		return &social.FriendPutInResp{}, err
	}

	reqUidInt, err := strconv.Atoi(in.ReqUid)
	if err != nil {
		return &social.FriendPutInResp{}, err
	}

	// 创建申请记录，使用中国时区
	chinaNow := utils.NowInChina()
	reqTime := time.Unix(in.ReqTime, 0).In(utils.ChinaLocation)

	// 如果已存在申请记录，更新记录让消息重新生效
	if friendReqs != nil {
		friendReqs.Status = 1        // 正常显示
		friendReqs.ReqMsg = in.ReqMsg
		friendReqs.ReqTime = reqTime
		friendReqs.HandleResult = 0  // 重置为待处理
		friendReqs.HandledAt = chinaNow
		friendReqs.ReceiverRead = 0  // 重置接收方已读，让对方重新收到通知
		friendReqs.SenderRead = 0    // 发起方也重置

		// 使用事务更新记录
		err = l.svcCtx.FriendRequestsModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			return l.svcCtx.FriendRequestsModel.Update(ctx, session, friendReqs)
		})
		if err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "update friendRequest err %v req %v ", err, in)
		}
		// 失效接收方的气泡缓存
		l.svcCtx.FriendRequestsModel.InvalidateCountCache(l.ctx, in.ReqUid)
		return &social.FriendPutInResp{}, nil
	}

	// 创建新申请记录
	_, err = l.svcCtx.FriendRequestsModel.Insert(l.ctx, &socialmodels.FriendRequests{
		UserId:       uint64(uidInt),
		ReqUid:       uint64(reqUidInt),
		ReqMsg:       in.ReqMsg,
		Status:       1, // 1-正常显示
		ReqTime:      reqTime,
		HandleResult: 0, // 0-待处理
		HandledAt:    chinaNow,
		ReceiverRead: 0, // 接收方未读
		SenderRead:   0, // 发起方未读
	})

	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "insert friendRequest err %v req %v ", err, in)
	}

	// 失效接收方的气泡缓存
	l.svcCtx.FriendRequestsModel.InvalidateCountCache(l.ctx, in.ReqUid)

	return &social.FriendPutInResp{}, nil
}
