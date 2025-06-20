package logic

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type ToggleLikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewToggleLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ToggleLikeLogic {
	return &ToggleLikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ToggleLike 点赞/取消点赞
func (l *ToggleLikeLogic) ToggleLike(in *trend.LikeToggleRequest) (*trend.LikeToggleResponse, error) {
	// todo: add your logic here and delete this line

	err := l.svcCtx.TrendAgree.AgreeInc(l.ctx, uint64(in.UserId), uint64(in.AuthorId), uint64(in.TrendId), int(in.LikeType))
	if err != nil {
		return nil, fmt.Errorf("用户: %d 点赞动态: %d 失败: %s", in.UserId, in.TrendId, err)
	}

	return &trend.LikeToggleResponse{
		Success: true,
	}, nil
}
