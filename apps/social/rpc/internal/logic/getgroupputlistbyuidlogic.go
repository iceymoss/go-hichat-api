package logic

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupPutListByUidLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupPutListByUidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupPutListByUidLogic {
	return &GetGroupPutListByUidLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetGroupPutListByUid 用户角度的群申请列表。
// 前端只调一次（class=2）后在本地按"申请人是否=我"拆分「我发起的/我收到的」，因此这里需返回两类的并集：
//   - 我发起的：req_id 属于 in.Ids（我提交的申请）
//   - 我收到的：group_id 属于「我作为群主/管理员管理的群」（别人申请进我管理的群）
//
// 仅保留 join_source ∈ (1主动申请, 3链接) 的真实申请，排除 2邀请 自动入群的噪声。
func (l *GetGroupPutListByUidLogic) GetGroupPutListByUid(in *social.GetGroupPutListByUidReq) (*social.GroupPutinListResp, error) {
	// 1. 汇总 in.Ids 作为群主/管理员管理的群（直接查 group_members 带 role_level，
	//    不用 GroupMembersModel.ListByUserId：它的列清单不含 role_level，拿不到角色）。
	var managed []string
	conn := l.svcCtx.DB
	if err := conn.WithContext(l.ctx).Table("group_members").
		Where("user_id in ? AND role_level >= ?", in.Ids, int(constants.ManagerGroupRoleLevel)).
		Pluck("group_id", &managed).Error; err != nil {
		return nil, err
	}

	// 2. 查询：req_id 属于我 OR group_id 属于我管理的群
	var rows []*socialmodels.GroupRequests
	query := conn.WithContext(l.ctx).Table("group_requests").Where("join_source in ?", []int{1, 3})
	if len(managed) > 0 {
		query = query.Where("req_id in ? OR group_id in ?", in.Ids, managed)
	} else {
		query = query.Where("req_id in ?", in.Ids)
	}
	if in.Type != "" && in.Type != "all" {
		query = query.Where("handle_result = ?", in.Type)
	}
	if err := query.Order("id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	requestIDs := make([]uint64, len(rows))
	for i, request := range rows {
		requestIDs[i] = uint64(request.Id)
	}
	receiptByKey := make(map[string]int)
	if len(requestIDs) > 0 && len(in.Ids) > 0 {
		var receipts []objects.SocialRequestReceipt
		if err := conn.WithContext(l.ctx).Where("request_type = ? AND request_id IN ? AND receiver_id IN ?", receiptTypeGroup, requestIDs, in.Ids).Find(&receipts).Error; err != nil {
			return nil, err
		}
		for _, receipt := range receipts {
			receiptByKey[fmt.Sprintf("%d:%s", receipt.RequestID, receipt.ReceiverID)] = receipt.IsRead
		}
	}

	respList := make([]*social.GroupRequests, 0, len(rows))
	for _, req := range rows {
		receiverRead := int32(req.ReceiverRead)
		for _, actor := range in.Ids {
			if read, ok := receiptByKey[fmt.Sprintf("%d:%s", req.Id, actor)]; ok {
				receiverRead = int32(read)
				break
			}
		}
		respList = append(respList, &social.GroupRequests{
			Id:               int32(req.Id),
			GroupId:          req.GroupId,
			ReqId:            req.ReqId,
			ReqMsg:           req.ReqMsg.String,
			ReqTime:          req.ReqTime.Time.Unix(),
			JoinSource:       int32(req.JoinSource.Int64),
			InviterUid:       req.InviterUserId.String,
			HandleUid:        req.HandleUserId.String,
			HandleResult:     int32(req.HandleResult.Int64),
			HandleResultTime: req.HandleTime.Unix(),
			ReceiverRead:     receiverRead,
		})
	}

	return &social.GroupPutinListResp{
		List: respList,
	}, nil
}
