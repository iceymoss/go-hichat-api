package logic

import (
	"context"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadLikesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadLikesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadLikesLogic {
	return &GetUnreadLikesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUnreadLikes 获取未读点赞通知
func (l *GetUnreadLikesLogic) GetUnreadLikes(in *trend.GetUnreadLikesRequest) (*trend.GetUnreadLikesResponse, error) {
	limit := 30
	agreeList, count, err := l.svcCtx.TrendAgree.GetUnReadAgree(l.ctx, int(in.UserId), []string{"*"}, int(in.LastId), limit)
	if err != nil {
		zLog.Error("get agree list failed", zap.Any("userId", in.UserId), zap.Error(err))
		return nil, err
	}

	list := make([]*trend.LikeInfo, 0, len(agreeList))
	for _, v := range agreeList {
		list = append(list, &trend.LikeInfo{
			Id:         v.Id,
			UserId:     uint32(v.Userid),
			TrendId:    uint32(v.TrendId),
			CreateTime: v.OpTime.Unix(),
		})
	}

	var lastId uint32
	if len(list) > 0 {
		lastId = uint32(list[len(list)-1].Id)
	}

	return &trend.GetUnreadLikesResponse{
		Likes:       list,
		TotalUnread: uint32(count),
		LastId:      lastId,
	}, nil
}
