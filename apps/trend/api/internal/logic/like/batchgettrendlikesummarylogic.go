package like

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type BatchGetTrendLikeSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewBatchGetTrendLikeSummaryLogic 批量获取点赞摘要
func NewBatchGetTrendLikeSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetTrendLikeSummaryLogic {
	return &BatchGetTrendLikeSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchGetTrendLikeSummaryLogic) BatchGetTrendLikeSummary(req *types.GetLikeListReq) (resp *types.GetLikeListResp, err error) {
	resp = &types.GetLikeListResp{LikeMap: map[uint64][]types.User{}}
	if len(req.TrendID) == 0 {
		return resp, nil
	}

	uid := utils.GetUser(l.ctx)

	trendIds := make([]uint32, 0, len(req.TrendID))
	for _, id := range req.TrendID {
		trendIds = append(trendIds, uint32(id))
	}

	summaryResp, err := l.svcCtx.Trend.BatchGetTrendLikeSummary(l.ctx, &trend.BatchTrendLikeSummaryRequest{
		UserId:   uint32(uid),
		TrendIds: trendIds,
	})
	if err != nil {
		zLog.Error("批量获取点赞摘要失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 收集所有点赞用户ID，一次性查询用户信息
	userIdSet := make(map[string]struct{})
	for _, summary := range summaryResp.Summaries {
		for _, like := range summary.PreviewLikes {
			userIdSet[strconv.Itoa(int(like.UserId))] = struct{}{}
		}
	}

	userBind := make(map[string]types.User, len(userIdSet))
	if len(userIdSet) > 0 {
		userIdList := make([]string, 0, len(userIdSet))
		for id := range userIdSet {
			userIdList = append(userIdList, id)
		}
		userInfo, uErr := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{Ids: userIdList})
		if uErr != nil {
			zLog.Error("获取点赞用户信息失败", zap.Any("ids", userIdList), zap.Error(uErr))
			return nil, uErr
		}
		for _, v := range userInfo.User {
			userBind[v.Id] = types.User{
				Id:       v.Id,
				Nickname: v.Nickname,
				Sex:      int(v.Sex),
				Avatar:   v.Avatar,
			}
		}
	}

	// 按动态ID聚合点赞用户列表
	likeMap := make(map[uint64][]types.User, len(summaryResp.Summaries))
	for trendId, summary := range summaryResp.Summaries {
		users := make([]types.User, 0, len(summary.PreviewLikes))
		for _, like := range summary.PreviewLikes {
			if u, ok := userBind[strconv.Itoa(int(like.UserId))]; ok {
				users = append(users, u)
			}
		}
		likeMap[uint64(trendId)] = users
	}

	resp.LikeMap = likeMap
	return resp, nil
}
