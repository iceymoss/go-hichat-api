package logic

import (
	"context"
	"database/sql"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrFriendReqBeforePass   = xerr.NewMsg("好友申请并已经通过")
	ErrFriendReqBeforeRefuse = xerr.NewMsg("好友申请已经被拒绝")
)

type FriendPutInHandleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInHandleLogic {
	return &FriendPutInHandleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInHandle 处理好友申请
func (l *FriendPutInHandleLogic) FriendPutInHandle(in *social.FriendPutInHandleReq) (*social.FriendPutInHandleResp, error) {
	// 获取好友申请记录
	firendReq, err := l.svcCtx.FriendRequestsModel.FindOne(l.ctx, uint64(in.FriendReqId), in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friendsRequest by friendReqid err %v req %v ", err,
			in.FriendReqId)
	}

	// 验证是否有处理
	switch constants.HandlerResult(firendReq.HandleResult) {
	case constants.PassHandlerResult: //已经通过直接返回
		return nil, errors.WithStack(ErrFriendReqBeforePass)
	case constants.RefuseHandlerResult: //已经拒绝直接返回
		return nil, errors.WithStack(ErrFriendReqBeforeRefuse)
	}

	firendReq.HandleResult = int(in.HandleResult)

	// 修改申请结果 -> 通过【建立两条好友关系记录】 -> 事务
	err = l.svcCtx.FriendRequestsModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.FriendRequestsModel.Update(l.ctx, session, firendReq); err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "update friend request err %v, req %v", err, firendReq)
		}

		if constants.HandlerResult(in.HandleResult) != constants.PassHandlerResult {
			return nil
		}

		friend1 := &socialmodels.Friends{
			UserId:    firendReq.UserId,
			FriendUid: firendReq.ReqUid,
			Remark:    string(firendReq.ReqUid),
			AddSource: 1,
			CreatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: false,
			},
		}

		friend2 := &socialmodels.Friends{
			UserId:    firendReq.ReqUid,
			FriendUid: firendReq.UserId,
			Remark:    string(firendReq.UserId),
			AddSource: 1,
			CreatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: false,
			},
		}

		_, err = l.svcCtx.FriendsModel.Insert(l.ctx, friend1)
		if err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "friends inserts err %v, req %v", err, friend1)
		}

		_, err = l.svcCtx.FriendsModel.Insert(l.ctx, friend2)
		if err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "friends inserts err %v, req %v", err, friend2)
		}

		//friends := []*socialmodels.Friends{
		//	{
		//		UserId:    firendReq.UserId,
		//		FriendUid: firendReq.ReqUid,
		//		Remark: sql.NullString{
		//			String: string(firendReq.ReqUid),
		//			Valid:  false,
		//		},
		//		AddSource: sql.NullInt64{
		//			Int64: 1,
		//			Valid: false,
		//		},
		//		CreatedAt: sql.NullTime{
		//			Time:  time.Now(),
		//			Valid: false,
		//		},
		//	}, {
		//		UserId:    firendReq.ReqUid,
		//		FriendUid: firendReq.UserId,
		//		Remark: sql.NullString{
		//			String: string(firendReq.ReqUid),
		//			Valid:  false,
		//		},
		//		AddSource: sql.NullInt64{
		//			Int64: 1,
		//			Valid: false,
		//		},
		//		CreatedAt: sql.NullTime{
		//			Time:  time.Now(),
		//			Valid: false,
		//		},
		//	},
		//}

		//_, err = l.svcCtx.FriendsModel.Inserts(l.ctx, session, friends...)
		//if err != nil {
		//	return errors.Wrapf(xerr.NewDBErr(), "friends inserts err %v, req %v", err, friends)
		//}
		return nil
	})

	return &social.FriendPutInHandleResp{}, err
}
