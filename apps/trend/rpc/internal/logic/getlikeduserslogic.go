package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetLikedUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLikedUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikedUsersLogic {
	return &GetLikedUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLikedUsers 获取点赞用户列表 (微信"查看全部点赞")
func (l *GetLikedUsersLogic) GetLikedUsers(in *trend.GetLikedUsersRequest) (*trend.GetLikedUsersResponse, error) {
	agree, count, err := l.svcCtx.TrendAgree.GetAgreeByTrendId(l.ctx, uint64(in.TrendId), []string{"id", "userid", "author_id", "op_time"}, int(in.Cursor), int(in.Limit))
	if err != nil {
		zLog.Error("GetLikedUsers.GetAgreeByTrendId: get agree list failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, err
	}

	users := make([]*trend.LikeInfo, 0, len(agree))
	for _, v := range agree {
		users = append(users, &trend.LikeInfo{
			Id:         v.Id,
			UserId:     uint32(v.Userid),
			TrendId:    in.TrendId,
			CreateTime: v.OpTime.Unix(),
		})
	}

	var lastId uint32
	if len(users) > 0 {
		lastId = uint32(users[len(users)-1].Id)
	}

	var hasMore bool
	if len(users) > int(in.Limit) {
		hasMore = true
	}

	return &trend.GetLikedUsersResponse{
		Users:      users,
		NextCursor: lastId,
		HasMore:    hasMore,
		Total:      uint32(count),
	}, nil
}
