package logic

import (
	"context"
	"math"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GroupPutinListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupPutinListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutinListLogic {
	return &GroupPutinListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GroupPutinList 获取用户加群申请列表
func (l *GroupPutinListLogic) GroupPutinList(in *social.GroupPutinListReq) (*social.GroupPutinListResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	if len(in.Type) == 0 {
		in.Type = []int32{0, 1, 2}
	}

	var list []*socialmodels.GroupRequests
	var listErr error
	if in.GetClass() == 1 {
		// 用户发起的申请
		list, listErr = l.svcCtx.ListReqByUser(l.ctx, actor)
		if listErr != nil {
			zLog.Error("GroupPutinList.ListReqByUser: ", zap.Any("groupId", in.GroupId), zap.Error(listErr))
			return nil, listErr
		}
	} else {
		// 只能管理员和群主可以看到，获取群信息
		member, findErr := l.svcCtx.GroupMembersModel.FindMemberByUid(l.ctx, in.GroupId, actor, []string{"role_level"})
		if findErr != nil {
			zLog.Error("GroupPutinList.FindOne: ", zap.Any("groupId", in.GroupId), zap.Any("userId", in.UserId), zap.Error(findErr))
			return nil, findErr
		}

		if int(member.RoleLevel) == 0 {
			return nil, errors.New("get group req need create_uid or manager_uid")
		}

		// 获取当前群的加入申请
		list, listErr = l.svcCtx.ListHandlerByGroup(l.ctx, in.GroupId, in.Type)
	}

	respList := make([]*social.GroupRequests, 0, len(list))
	ids := make([]uint64, 0, len(list))
	for _, v := range list {
		if v.Id > 0 {
			ids = append(ids, uint64(v.Id))
		}
	}
	canonical := map[uint64]objects.GroupRequest{}
	if len(ids) > 0 {
		var rows []objects.GroupRequest
		if findErr := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", ids).Find(&rows).Error; findErr != nil {
			return nil, findErr
		}
		for _, row := range rows {
			canonical[row.ID] = row
		}
	}
	receipts := map[uint64]objects.SocialRequestReceipt{}
	if len(ids) > 0 {
		kind := receiptKindApply
		if in.GetClass() == 1 {
			kind = receiptKindResult
		}
		var rows []objects.SocialRequestReceipt
		if findErr := l.svcCtx.DB.WithContext(l.ctx).Where("request_type=? AND request_id IN ? AND receiver_id=? AND receipt_kind=?", receiptTypeGroup, ids, actor, kind).Find(&rows).Error; findErr != nil {
			return nil, findErr
		}
		for _, row := range rows {
			receipts[row.RequestID] = row
		}
	}
	for _, v := range list {
		id := uint64(v.Id)
		row := canonical[id]
		legacyID := int32(0)
		if id <= math.MaxInt32 {
			legacyID = int32(id)
		}
		read := int32(v.ReceiverRead)
		actionable := false
		if receipt, ok := receipts[id]; ok {
			read = int32(receipt.IsRead)
			actionable = receipt.IsActionable == 1
		}
		sourceInvitation := uint64(0)
		if row.SourceInvitationID != nil {
			sourceInvitation = *row.SourceInvitationID
		}
		actual := int32(0)
		if row.ActualJoinSource != nil {
			actual = int32(*row.ActualJoinSource)
		}
		handleUID := v.HandleUserId.String
		if in.GetClass() == 1 {
			handleUID = ""
		}
		respList = append(respList, &social.GroupRequests{
			Id:               legacyID,
			RequestId:        id,
			GroupId:          v.GroupId,
			ReqId:            v.ReqId,
			ReqMsg:           v.ReqMsg.String,
			ReqTime:          v.ReqTime.Time.Unix(),
			JoinSource:       int32(v.JoinSource.Int64),
			InviterUid:       v.InviterUserId.String,
			HandleUid:        handleUID,
			HandleResult:     int32(v.HandleResult.Int64),
			HandleResultTime: v.HandleTime.Unix(),
			ApplicantUid:     v.ReqId, HandleMsg: row.HandleMsg, InvalidReason: row.InvalidReason, ActualJoinSource: actual, SourceType: int32(row.SourceType), SourceInvitationId: sourceInvitation, ReadState: read, ReceiverRead: read, Actionable: actionable,
		})
	}
	return &social.GroupPutinListResp{
		List:  respList,
		Total: int64(len(respList)),
	}, listErr
}
