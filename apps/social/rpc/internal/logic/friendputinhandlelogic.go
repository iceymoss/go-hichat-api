package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/utils"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	case constants.IgnoreHandlerResult: //已经忽略直接返回
		return nil, errors.WithStack(ErrFriendReqBeforeRefuse) // 使用相同的错误，因为忽略和拒绝都是不可再处理
	}

	// 获取用户信息
	friendRsep, err := l.svcCtx.User.GetUserById(l.ctx, &user.GetUserByIdRequest{
		Id: in.UserId,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "get user by id err %v, req %v", err, in.UserId)
	}

	firendReq.HandleResult = int(in.HandleResult)
	// 保存处理附言
	if in.HandleMsg != "" {
		firendReq.HandleMsg = in.HandleMsg
	}
	// 处理后：发起方需要收到结果通知，重置 sender_read=0
	firendReq.SenderRead = 0
	// 使用中国时区更新处理时间
	chinaNow := utils.NowInChina()
	firendReq.HandledAt = chinaNow

	// 获取申请人的用户信息（用于设置默认备注）
	applicantResp, _ := l.svcCtx.User.GetUserById(l.ctx, &user.GetUserByIdRequest{
		Id: fmt.Sprint(firendReq.UserId),
	})

	// 修改申请结果 -> 通过【建立两条好友关系记录】 -> 事务
	err = l.svcCtx.FriendRequestsModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.FriendRequestsModel.Update(l.ctx, session, firendReq); err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "update friend request err %v, req %v", err, firendReq)
		}

		// 将同一对用户之间其他待处理的申请标记为已忽略
		if err := l.svcCtx.FriendRequestsModel.IgnoreOtherPending(
			l.ctx, uint64(in.FriendReqId), firendReq.UserId, firendReq.ReqUid,
		); err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "ignore other pending err %v", err)
		}

		if constants.HandlerResult(in.HandleResult) != constants.PassHandlerResult {
			return nil
		}

		// 申请人好友列表中添加处理人
		// 申请人发起申请时若预设了备注，则用预设备注；否则用处理人昵称作默认
		applicantRemark := firendReq.Remark
		if applicantRemark == "" {
			applicantRemark = friendRsep.User.Nickname
		}
		friend1 := &socialmodels.Friends{
			UserId:    firendReq.UserId,
			FriendUid: firendReq.ReqUid,
			Remark:    applicantRemark,
			AddSource: 1,
			CreatedAt: sql.NullTime{
				Time:  chinaNow,
				Valid: true,
			},
		}

		// 处理标签
		if len(in.Tags) > 0 {
			tagBytes, _ := json.Marshal(in.Tags)
			friend1.FriendTags = sql.NullString{String: string(tagBytes), Valid: true}
		}

		// 处理人好友列表中添加申请人
		// remark 是处理人给申请人设的备注；如果没设，用申请人的昵称作默认
		remark := in.Remark
		if remark == "" && applicantResp != nil && applicantResp.User != nil {
			remark = applicantResp.User.Nickname
		}
		friend2 := &socialmodels.Friends{
			UserId:    firendReq.ReqUid,
			FriendUid: firendReq.UserId,
			Remark:    remark,
			AddSource: 1,
			CreatedAt: sql.NullTime{
				Time:  chinaNow,
				Valid: true,
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

	// 失效双方气泡缓存：发起方收到结果通知，接收方数量也变了
	l.svcCtx.FriendRequestsModel.InvalidateCountCache(l.ctx,
		fmt.Sprint(firendReq.UserId), fmt.Sprint(firendReq.ReqUid))

	return &social.FriendPutInHandleResp{}, err
}
