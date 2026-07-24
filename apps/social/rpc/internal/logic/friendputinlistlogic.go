package logic

import (
	"context"
	"math"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FriendPutInListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInListLogic(ctx context.Context, s *svc.ServiceContext) *FriendPutInListLogic {
	return &FriendPutInListLogic{ctx: ctx, svcCtx: s, Logger: logx.WithContext(ctx)}
}

func (l *FriendPutInListLogic) FriendPutInList(in *social.FriendPutInListReq) (*social.FriendPutInListResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	if in.Class != "0" && in.Class != "1" {
		return nil, status.Error(codes.InvalidArgument, "invalid friend request list scope")
	}
	page, size := int(in.Page), int(in.Size)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&objects.FriendRequest{}).Where("status <> ?", 0)
	hidden := l.svcCtx.DB.Model(&objects.SocialRequestReceipt{}).Select("request_id").
		Where("request_type = ? AND receiver_id = ? AND result = ?", receiptTypeFriend, actor, receiptInvalidated)
	query = query.Where("id NOT IN (?)", hidden)
	if in.Class == "0" {
		query = query.Where("user_id = ?", actor)
	} else {
		query = query.Where("req_uid = ?", actor)
	}
	if in.Status != nil {
		query = query.Where("handle_result = ?", *in.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to count friend requests")
	}
	var rows []objects.FriendRequest
	if err := query.Order("req_time DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to list friend requests")
	}
	ids := make([]uint64, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	type receiptState struct{ Read, Actionable int }
	receipts := map[uint64]receiptState{}
	if len(ids) > 0 {
		kind := receiptKindApply
		if in.Class == "0" {
			kind = receiptKindResult
		}
		var rs []objects.SocialRequestReceipt
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("request_type=? AND request_id IN ? AND receiver_id=? AND receipt_kind=?", receiptTypeFriend, ids, actor, kind).Find(&rs).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to list friend request receipts")
		}
		for _, r := range rs {
			receipts[r.RequestID] = receiptState{Read: r.IsRead, Actionable: r.IsActionable}
		}
	}
	list := make([]*social.FriendRequests, 0, len(rows))
	for _, r := range rows {
		read := r.ReceiverRead
		peer := strconv.FormatUint(r.UserID, 10)
		if in.Class == "0" {
			read = r.SenderRead
			peer = strconv.FormatUint(r.ReqUID, 10)
		}
		handled := int64(0)
		if r.HandledAt != nil {
			handled = r.HandledAt.Unix()
		}
		result, state := 0, 1
		if r.HandleResult != nil {
			result = *r.HandleResult
		}
		if r.Status != nil {
			state = *r.Status
		}
		actionable := in.Class == "1" && result == receiptPending
		if v, ok := receipts[r.ID]; ok {
			read = v.Read
			actionable = v.Actionable == 1
		}
		legacyID := int32(0)
		if r.ID <= math.MaxInt32 {
			legacyID = int32(r.ID)
		}
		list = append(list, &social.FriendRequests{Id: legacyID, RequestId: r.ID, UserId: strconv.FormatUint(r.UserID, 10), ReqUid: strconv.FormatUint(r.ReqUID, 10), PeerUid: peer, ReqMsg: r.ReqMsg, ReqTime: r.ReqTime.Unix(), HandleResult: int32(result), Status: int32(state), ReadState: int32(read), HandleMsg: r.HandleMsg, HandledAt: handled, Actionable: actionable})
	}
	return &social.FriendPutInListResp{List: list, Total: total}, nil
}
