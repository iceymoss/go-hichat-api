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

	// 获取动态
	trendDetail, err := l.svcCtx.Trend.GetTrendDetail(l.ctx, &trend.GetTrendDetailRequest{TrendId: int64(req.TrendID)})
	if err != nil {
		zLog.Error("获取动态详情失败", zap.Error(err))
		return nil, err
	}

	authorID := strconv.Itoa(int(trendDetail.Trend.UserId))
	currentUID := strconv.Itoa(uid)

	// 可见性判定:
	// scope = 1 仅自己可见 → uid == author
	// scope = 2 仅好友可见 → uid == author 或 uid ∈ author 的好友
	// scope = 3 所有人可见 → 直接放行
	scope := trendDetail.Trend.Scope
	if authorID != currentUID && scope != 3 {
		if scope == 1 {
			return nil, errors.New("该动态仅作者可见")
		}
		friendList, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{UserId: currentUID})
		if err != nil {
			zLog.Error("获取好友列表失败", zap.Error(err))
			return nil, err
		}
		isFriend := false
		for _, v := range friendList.List {
			if v.FriendUid == authorID {
				isFriend = true
				break
			}
		}
		if !isFriend {
			return nil, errors.New("该动态不是您的好友发布的")
		}
	}

	// 一次性拉作者 + 被 @ 的用户信息
	userIdList := []string{authorID}
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
	if authorID != currentUID {
		detail.IsTop = 0
	}
	for _, v := range userInfo.User {
		if v.Id == authorID {
			detail.User = &types.User{
				Id:       v.Id,
				Nickname: v.Nickname,
				Sex:      int(v.Sex),
				Avatar:   v.Avatar,
			}
			continue
		}
		detail.AtUserIds = append(detail.AtUserIds, &types.User{
			Id:       v.Id,
			Nickname: v.Nickname,
			Sex:      int(v.Sex),
			Avatar:   v.Avatar,
		})
	}

	resp = &types.GetTrendDetailResponse{Trend: detail}

	return
}
