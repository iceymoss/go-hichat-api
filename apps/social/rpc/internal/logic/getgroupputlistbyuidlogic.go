package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

func (l *GetGroupPutListByUidLogic) GetGroupPutListByUid(in *social.GetGroupPutListByUidReq) (*social.GroupPutinListResp, error) {
	// 获取用户申请列表
	groupReqList, err := l.svcCtx.GroupRequestsModel.GetGroupReqListByUid(l.ctx, in.Ids, in.Class, in.Type)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &social.GroupPutinListResp{
				List: nil,
			}, nil
		}
	}

	var respList []*social.GroupRequests
	for _, req := range groupReqList {
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
		})
	}

	return &social.GroupPutinListResp{
		List: respList,
	}, nil
}
