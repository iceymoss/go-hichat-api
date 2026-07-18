package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/utils"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInListLogic {
	return &FriendPutInListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInList 获取未处理的好友申请列表,或者获取我发起的申请好友列表
func (l *FriendPutInListLogic) FriendPutInList(in *social.FriendPutInListReq) (*social.FriendPutInListResp, error) {
	friendReqList, err := l.svcCtx.FriendRequestsModel.ListFilterHandler(l.ctx, in.UserId, in.Type, in.Class)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find list friend req err %v req %v", err, in.UserId)
	}
	requestIDs := make([]uint64, len(friendReqList))
	for i, request := range friendReqList {
		requestIDs[i] = uint64(request.Id)
	}
	kind := receiptKindApply
	if in.Class == "0" {
		kind = receiptKindResult
	}
	var receipts []objects.SocialRequestReceipt
	if len(requestIDs) > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("request_type = ? AND request_id IN ? AND receiver_id = ? AND receipt_kind = ?", receiptTypeFriend, requestIDs, in.UserId, kind).Find(&receipts).Error; err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "find friend request receipts err %v", err)
		}
	}
	readByID := make(map[uint64]int, len(receipts))
	for _, receipt := range receipts {
		readByID[receipt.RequestID] = receipt.IsRead
	}

	resp := make([]*social.FriendRequests, 0, len(friendReqList))
	for _, v := range friendReqList {
		// 将时间转换为中国时区的Unix时间戳
		reqTimeUnix := utils.TimeToChinaUnix(v.ReqTime)

		// 根据 status 控制消息显示：status=2（忽略）时不返回消息
		reqMsg := v.ReqMsg
		if v.Status == 2 {
			reqMsg = "" // status=2 时不显示消息
		}

		// read_state 返回当前用户视角的已读状态
		// class="1"(我收到的) → 用 receiver_read
		// class="0"(我发出的) → 用 sender_read
		readState := v.ReceiverRead
		if in.Class == "0" {
			readState = v.SenderRead
		}
		if receiptRead, ok := readByID[uint64(v.Id)]; ok {
			readState = receiptRead
		}

		resp = append(resp, &social.FriendRequests{
			Id:           int32(v.Id),
			UserId:       strconv.Itoa(int(v.UserId)),
			ReqUid:       strconv.Itoa(int(v.ReqUid)),
			ReqMsg:       reqMsg,
			Status:       int32(v.Status), // 消息状态（0:已删除 1:正常显示 2:忽略不显示）
			ReqTime:      reqTimeUnix,
			HandleResult: int32(v.HandleResult), // 0-待处理, 1-已同意, 2-已拒绝
			ReadState:    int32(readState),
		})
	}

	return &social.FriendPutInListResp{
		List: resp,
	}, nil
}
