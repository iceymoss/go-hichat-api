package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
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
	groupReqs, err := l.svcCtx.GroupRequestsModel.ListNoHandler(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list group req err %v req %v", err, in.GroupId)
	}

	respList := make([]*social.GroupRequests, 0, len(groupReqs))
	for _, v := range groupReqs {
		respList = append(respList, &social.GroupRequests{
			Id:           int32(v.Id),
			GroupId:      v.GroupId,
			ReqId:        v.ReqId,
			ReqMsg:       v.ReqMsg.String,
			ReqTime:      v.ReqTime.Time.Unix(),
			JoinSource:   int32(v.JoinSource.Int64),
			InviterUid:   v.InviterUserId.String,
			HandleUid:    v.HandleUserId.String,
			HandleResult: int32(v.HandleResult.Int64),
		})
	}
	return &social.GroupPutinListResp{
		List: respList,
	}, nil
}
