package logic

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetGroupPutListByUidLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupPutListByUidLogic(ctx context.Context, s *svc.ServiceContext) *GetGroupPutListByUidLogic {
	return &GetGroupPutListByUidLogic{ctx: ctx, svcCtx: s, Logger: logx.WithContext(ctx)}
}

func (l *GetGroupPutListByUidLogic) GetGroupPutListByUid(in *social.GetGroupPutListByUidReq) (*social.GroupPutinListResp, error) {
	actor, err := validateScopedActor(in.ActorUid)
	if err != nil {
		return nil, err
	}
	if len(in.Ids) == 0 || (in.Class != "1" && in.Class != "2") {
		return nil, status.Error(codes.InvalidArgument, "invalid group request list scope")
	}
	for _, id := range in.Ids {
		if id != actor {
			return nil, status.Error(codes.PermissionDenied, "actor uid does not match requested scope")
		}
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
	q := l.svcCtx.DB.WithContext(l.ctx).Model(&objects.GroupRequest{})
	if in.Class == "1" {
		q = q.Where("req_id = ?", actor)
	} else {
		q = q.Where("EXISTS (SELECT 1 FROM social_request_receipts srr WHERE srr.request_type = ? AND srr.request_id = group_requests.id AND srr.receiver_id = ? AND srr.receipt_kind = ?)", receiptTypeGroup, actor, receiptKindApply)
	}
	if in.Status != nil {
		q = q.Where("handle_result = ?", *in.Status)
	} else if in.Type != "" && in.Type != "all" {
		v, e := strconv.Atoi(in.Type)
		if e != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid group request status")
		}
		q = q.Where("handle_result = ?", v)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to count group requests")
	}
	var rows []objects.GroupRequest
	if err := q.Order("req_time DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to list group requests")
	}
	ids := make([]uint64, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	receipts := map[string]objects.SocialRequestReceipt{}
	if len(ids) > 0 {
		var rs []objects.SocialRequestReceipt
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("request_type=? AND request_id IN ? AND receiver_id=? AND receipt_kind IN ?", receiptTypeGroup, ids, actor, []string{receiptKindApply, receiptKindResult}).Find(&rs).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to list group request receipts")
		}
		for _, r := range rs {
			receipts[fmt.Sprintf("%d:%s", r.RequestID, r.ReceiptKind)] = r
		}
	}
	list := make([]*social.GroupRequests, 0, len(rows))
	for _, r := range rows {
		read := int32(r.ReceiverRead)
		action := false
		kind := receiptKindApply
		if r.ReqID == actor {
			kind = receiptKindResult
		}
		if receipt, ok := receipts[fmt.Sprintf("%d:%s", r.ID, kind)]; ok {
			read = int32(receipt.IsRead)
			action = receipt.IsActionable == 1
		}
		reqTime, handleTime := int64(0), int64(0)
		if r.ReqTime != nil {
			reqTime = r.ReqTime.Unix()
		}
		if r.HandleTime != nil {
			handleTime = r.HandleTime.Unix()
		}
		join, result, actual := 0, 0, 0
		if r.JoinSource != nil {
			join = *r.JoinSource
		}
		if r.HandleResult != nil {
			result = *r.HandleResult
		}
		if r.ActualJoinSource != nil {
			actual = *r.ActualJoinSource
		}
		inviter, handler := "", ""
		if r.InviterUserID != nil {
			inviter = fmt.Sprint(*r.InviterUserID)
		}
		if r.HandleUserID != nil {
			handler = fmt.Sprint(*r.HandleUserID)
		}
		if in.Class == "1" {
			handler = ""
		}
		sourceInvitation := uint64(0)
		if r.SourceInvitationID != nil {
			sourceInvitation = *r.SourceInvitationID
		}
		legacyID := int32(0)
		if r.ID <= math.MaxInt32 {
			legacyID = int32(r.ID)
		}
		list = append(list, &social.GroupRequests{Id: legacyID, RequestId: r.ID, GroupId: fmt.Sprint(r.GroupID), ReqId: r.ReqID, ApplicantUid: r.ReqID, ReqMsg: r.ReqMsg, ReqTime: reqTime, JoinSource: int32(join), InviterUid: inviter, HandleUid: handler, HandleResult: int32(result), HandleResultTime: handleTime, ReceiverRead: read, ReadState: read, Actionable: action, HandleMsg: r.HandleMsg, InvalidReason: r.InvalidReason, ActualJoinSource: int32(actual), SourceType: int32(r.SourceType), SourceInvitationId: sourceInvitation})
	}
	return &social.GroupPutinListResp{List: list, Total: total}, nil
}
