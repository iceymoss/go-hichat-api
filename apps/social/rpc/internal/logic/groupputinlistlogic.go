package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"

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
	if len(in.Type) == 0 {
		in.Type = []int32{0, 1, 2}
	}

	var list []*socialmodels.GroupRequests
	var err error
	if in.GetClass() == 1 {
		// 用户发起的申请
		list, err = l.svcCtx.ListReqByUser(l.ctx, in.UserId)
		if err != nil {
			zLog.Error("GroupPutinList.ListReqByUser: ", zap.Any("groupId", in.GroupId), zap.Error(err))
			return nil, err
		}
	} else {
		// 只能管理员和群主可以看到，获取群信息
		member, findErr := l.svcCtx.GroupMembersModel.FindMemberByUid(l.ctx, in.GroupId, in.UserId, []string{"role_level"})
		if findErr != nil {
			zLog.Error("GroupPutinList.FindOne: ", zap.Any("groupId", in.GroupId), zap.Any("userId", in.UserId), zap.Error(findErr))
			return nil, findErr
		}

		if int(member.RoleLevel) == 0 {
			return nil, errors.New("get group req need create_uid or manager_uid")
		}

		// 获取当前群的加入申请
		list, err = l.svcCtx.ListHandlerByGroup(l.ctx, in.GroupId, in.Type)
	}

	respList := make([]*social.GroupRequests, 0, len(list))
	for _, v := range list {
		respList = append(respList, &social.GroupRequests{
			Id:               int32(v.Id),
			GroupId:          v.GroupId,
			ReqId:            v.ReqId,
			ReqMsg:           v.ReqMsg.String,
			ReqTime:          v.ReqTime.Time.Unix(),
			JoinSource:       int32(v.JoinSource.Int64),
			InviterUid:       v.InviterUserId.String,
			HandleUid:        v.HandleUserId.String,
			HandleResult:     int32(v.HandleResult.Int64),
			HandleResultTime: v.HandleTime.Unix(),
		})
	}
	return &social.GroupPutinListResp{
		List: respList,
	}, err
}
