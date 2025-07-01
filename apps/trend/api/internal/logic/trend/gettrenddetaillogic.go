package trend

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetTrendDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetTrendDetailLogic 获取动态详情
func NewGetTrendDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDetailLogic {
	return &GetTrendDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTrendDetailLogic) GetTrendDetail(req *types.GetTrendDetailRequest) (resp *types.GetTrendDetailResponse, err error) {
	uid := utils.GetUser(l.ctx)

	if req.TrendID == 0 {
		return nil, errors.New("动态id不能为0")
	}

	// 获取用户好友列表
	friendList, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{UserId: strconv.Itoa(uid)})
	if err != nil {
		zLog.Error("获取好友列表失败", zap.Error(err))
		return nil, err
	}

	// 获取动态
	trendDetail, err := l.svcCtx.Trend.GetTrendDetail(l.ctx, &trend.GetTrendDetailRequest{TrendId: int64(req.TrendID)})
	if err != nil {
		zLog.Error("获取动态详情失败", zap.Error(err))
		return nil, err
	}

	for _, v := range friendList.List {
		if v.FriendUid == strconv.Itoa(int(trendDetail.Trend.UserId)) || v.FriendUid == strconv.Itoa(uid) {
			break
		}
		return nil, errors.New("该动态不是您的好友发布的")
	}

	userIdList := []string{strconv.Itoa(uid)}
	for _, v := range trendDetail.Trend.AtUserIds {
		userIdList = append(userIdList, strconv.Itoa(int(v)))
	}

	userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: userIdList,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Error(err))
		return nil, err
	}

	detail := trendRpc2api(trendDetail.Trend)
	for _, v := range userInfo.User {
		if strconv.Itoa(uid) != v.Id {
			detail.AtUserIds = append(detail.AtUserIds, &types.User{
				Id:       v.Id,
				Nickname: v.Nickname,
				Sex:      int(v.Sex),
				Avatar:   v.Avatar,
			})
			continue
		}
		detail.User = &types.User{
			Id:       v.Id,
			Nickname: v.Nickname,
			Sex:      int(v.Sex),
			Avatar:   v.Avatar,
		}
	}

	resp = &types.GetTrendDetailResponse{Trend: detail}

	return
}
