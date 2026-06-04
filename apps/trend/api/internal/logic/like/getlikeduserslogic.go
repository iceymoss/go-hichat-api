package like

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

type GetLikedUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLikedUsersLogic 获取点赞用户列表
func NewGetLikedUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikedUsersLogic {
	return &GetLikedUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLikedUsersLogic) GetLikedUsers(req *types.GetLikedUsersRequest) (resp *types.GetLikedUsersResponse, err error) {
	likeUser, err := l.svcCtx.Trend.GetLikedUsers(l.ctx, &trend.GetLikedUsersRequest{
		TrendId: req.TrendID,
		Cursor:  req.Cursor,
		Limit:   req.Limit,
	})
	if err != nil {
		zLog.Error("获取动态点赞列表失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	userIdList := make([]string, 0, len(likeUser.Users))
	for _, v := range likeUser.Users {
		userIdList = append(userIdList, strconv.Itoa(int(v.UserId)))
	}

	// 获取用户信息
	userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: userIdList,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Any("userIdList", userIdList), zap.Error(err))
		return nil, err
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

	list := make([]*types.User, 0, len(likeUser.Users))
	for _, v := range likeUser.Users {
		list = append(list, userBind[strconv.Itoa(int(v.UserId))])
	}

	resp = &types.GetLikedUsersResponse{
		Users:      list,
		NextCursor: int(likeUser.NextCursor),
		HasMore:    likeUser.HasMore,
		Total:      int(likeUser.Total),
	}

	return
}
