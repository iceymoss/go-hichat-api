package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/utils"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

	isPass := constants.HandlerResult(in.HandleResult) == constants.PassHandlerResult

	// reqUpdate 在调用方事务里更新本条申请状态，并把同一对用户之间其他待处理申请标记为已忽略。
	// 复刻原 sqlx Update + IgnoreOtherPending 的列写入，改走 GORM 以便与 outbox 同事务原子提交。
	reqUpdate := func(tx *gorm.DB) error {
		updates := map[string]any{
			"handle_result": firendReq.HandleResult,
			"sender_read":   firendReq.SenderRead,
			"handled_at":    firendReq.HandledAt,
		}
		if in.HandleMsg != "" {
			updates["handle_msg"] = firendReq.HandleMsg
		}
		if err := tx.Table("friend_requests").Where("id = ?", firendReq.Id).Updates(updates).Error; err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "update friend request err %v, req %v", err, firendReq)
		}
		// 将同一对用户之间其他待处理的申请标记为已忽略
		if err := tx.Table("friend_requests").
			Where("id != ?", firendReq.Id).
			Where("handle_result = ?", 0).
			Where("status = ?", 1).
			Where("(user_id = ? AND req_uid = ?) OR (user_id = ? AND req_uid = ?)",
				firendReq.UserId, firendReq.ReqUid, firendReq.ReqUid, firendReq.UserId).
			Updates(map[string]any{"handle_result": 3}).Error; err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "ignore other pending err %v", err)
		}
		return nil
	}

	if isPass {
		// 申请人好友列表中添加处理人：备注用申请预设，否则用处理人昵称作默认
		applicantRemark := firendReq.Remark
		if applicantRemark == "" {
			applicantRemark = friendRsep.User.Nickname
		}
		friend1 := &socialmodels.Friends{
			UserId:    firendReq.UserId,
			FriendUid: firendReq.ReqUid,
			Remark:    applicantRemark,
			AddSource: 1,
			CreatedAt: sql.NullTime{Time: chinaNow, Valid: true},
		}
		if len(in.Tags) > 0 {
			tagBytes, _ := json.Marshal(in.Tags)
			friend1.FriendTags = sql.NullString{String: string(tagBytes), Valid: true}
		}

		// 处理人好友列表中添加申请人：备注用处理人输入，否则用申请人昵称作默认
		remark := in.Remark
		if remark == "" && applicantResp != nil && applicantResp.User != nil {
			remark = applicantResp.User.Nickname
		}
		friend2 := &socialmodels.Friends{
			UserId:    firendReq.ReqUid,
			FriendUid: firendReq.UserId,
			Remark:    remark,
			AddSource: 1,
			CreatedAt: sql.NullTime{Time: chinaNow, Valid: true},
		}

		// 双向建立好友关系 + 同事务写 outbox（好友新增事件），提交后 best-effort 同步关系缓存
		err = emitRelationChange(l.ctx, l.svcCtx, constants.RelationEventFriendAdded, "",
			&mq.RelationChangeTransfer{
				FriendA:    fmt.Sprint(firendReq.UserId),
				FriendB:    fmt.Sprint(firendReq.ReqUid),
				OperatorId: in.UserId,
			},
			func(tx *gorm.DB) error {
				if err := reqUpdate(tx); err != nil {
					return err
				}
				if err := tx.Table("friends").Create(friend1).Error; err != nil {
					return errors.Wrapf(xerr.NewDBErr(), "friends insert err %v, req %v", err, friend1)
				}
				if err := tx.Table("friends").Create(friend2).Error; err != nil {
					return errors.Wrapf(xerr.NewDBErr(), "friends insert err %v, req %v", err, friend2)
				}
				return nil
			})
	} else {
		// 拒绝/忽略：仅更新申请状态，不建立好友关系、不发关系事件
		tx := db.GetMysqlConn(db.MYSQL_DB_HICHAT2).Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}
		if e := reqUpdate(tx); e != nil {
			tx.Rollback()
			err = e
		} else {
			err = tx.Commit().Error
		}
	}

	// 失效双方气泡缓存：发起方收到结果通知，接收方数量也变了
	l.svcCtx.FriendRequestsModel.InvalidateCountCache(l.ctx,
		fmt.Sprint(firendReq.UserId), fmt.Sprint(firendReq.ReqUid))

	// 公共通知：申请人收到处理结果（通过/拒绝；忽略不通知）。仅业务成功时发。
	if err == nil {
		var notifyType string
		switch constants.HandlerResult(in.HandleResult) {
		case constants.PassHandlerResult:
			notifyType = NotifyFriendAccept
		case constants.RefuseHandlerResult:
			notifyType = NotifyFriendReject
		}
		if notifyType != "" {
			emitCommonNotify(l.ctx, l.svcCtx, &mq.CommonNotify{
				NotifyType: notifyType,
				ReceiverId: fmt.Sprint(firendReq.UserId),
				ActorId:    in.UserId,
				BizId:      fmt.Sprintf("friend.handle:%d:%d", firendReq.Id, in.HandleResult),
				Content:    in.HandleMsg,
			})
		}
	}

	return &social.FriendPutInHandleResp{}, err
}
