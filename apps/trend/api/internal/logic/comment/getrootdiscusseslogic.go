package comment

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRootDiscussesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetRootDiscussesLogic 获取一级评论
func NewGetRootDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootDiscussesLogic {
	return &GetRootDiscussesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRootDiscussesLogic) GetRootDiscusses(req *types.GetDiscussesReq) (resp *types.DiscussesListResp, err error) {
	discusses, err := l.svcCtx.Trend.GetRootDiscusses(l.ctx, &trend.GetDiscussesReq{
		TrendId: req.TrendID,
		Pagination: &trend.Pagination{
			LastId:   int32(req.LastID),
			LastTime: int64(req.LastTime),
		},
	})
	if err != nil {
		zLog.Error("获取一级评论失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	userIds := make([]string, 0)

	list := make([]*types.Discuss, 0, len(discusses.Discusses))
	for _, v := range discusses.Discusses {
		list = append(list, discuss2Resp(v))
		userIds = append(userIds, strconv.Itoa(int(v.UserId)), strconv.Itoa(int(v.Replyer)))
		for _, atUid := range v.AtUserIds {
			userIds = append(userIds, strconv.Itoa(int(atUid)))
		}
	}

	// 获取用户信息
	userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: userIds,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Any("userIds", userIds), zap.Error(err))
		return nil, nil
	}

	userBind := make(map[string]*types.User, len(userInfo.User))
	for _, v := range userInfo.User {
		userBind[v.Id] = &types.User{
			Id:       v.Id,
			Nickname: v.Nickname,
			Sex:      int(v.Sex),
			Avatar:   v.Avatar,
		}
	}

	for i, v := range discusses.Discusses {
		list[i].Replyer = userBind[strconv.Itoa(int(v.UserId))]
		list[i].User = userBind[strconv.Itoa(int(v.UserId))]
		for _, atUid := range v.AtUserIds {
			list[i].AtUser = append(list[i].AtUser, userBind[strconv.Itoa(int(atUid))])
		}
	}

	resp = &types.DiscussesListResp{
		DiscussesJSON: list,
		LastID:        int(discusses.Pagination.LastId),
		LastTime:      int(discusses.Pagination.LastTime),
		Total:         int(discusses.Total),
	}

	return
}
