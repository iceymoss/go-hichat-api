package comment

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetChildDiscussesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetChildDiscussesLogic 获取子评论列表
func NewGetChildDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildDiscussesLogic {
	return &GetChildDiscussesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetChildDiscusses 获取子评论
func (l *GetChildDiscussesLogic) GetChildDiscusses(req *types.GetChildDiscussesReq) (resp *types.DiscussesListResp, err error) {
	child, err := l.svcCtx.Trend.GetChildDiscusses(l.ctx, &trend.GetChildDiscussesReq{
		Father: req.Father,
		Pagination: &trend.Pagination{
			LastId:   int32(req.LastID),
			LastTime: int64(req.LastTime),
		},
	})
	if err != nil {
		zLog.Error("获取子评论失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	userIds := make([]string, 0)

	list := make([]*types.Discuss, 0, len(child.Discusses))
	for _, v := range child.Discusses {
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

	for i, v := range child.Discusses {
		list[i].Replyer = userBind[strconv.Itoa(int(v.UserId))]
		list[i].User = userBind[strconv.Itoa(int(v.UserId))]
		for _, atUid := range v.AtUserIds {
			list[i].AtUser = append(list[i].AtUser, userBind[strconv.Itoa(int(atUid))])
		}
	}

	resp = &types.DiscussesListResp{
		DiscussesJSON: list,
		LastID:        int(child.Pagination.LastId),
		LastTime:      int(child.Pagination.LastTime),
		Total:         int(child.Total),
	}

	return
}

func discuss2Resp(child *trend.Discuss) *types.Discuss {
	return &types.Discuss{
		Id:           int64(child.Id),
		TrendId:      int64(child.TrendId),
		RootId:       int64(child.RootId),
		Father:       int64(child.Father),
		Level:        int(child.Level),
		Content:      child.Content,
		AgreeCount:   int64(child.AgreeCount),
		DiscussCount: int64(child.DiscussCount),
		State:        int(child.State),
		Read:         child.Read,
		CreateTime:   child.CreateTime,
		UpdateTime:   child.UpdateTime,
		User:         &types.User{},
		Replyer:      &types.User{},
	}
}
