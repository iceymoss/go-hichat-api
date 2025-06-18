package logic

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type DeleteDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDiscussLogic {
	return &DeleteDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteDiscuss 删除评论（支持任意级别评论删除）
func (l *DeleteDiscussLogic) DeleteDiscuss(in *trend.DeleteDiscussReq) (*trend.DeleteDiscussResp, error) {
	discus, err := l.svcCtx.TrendDiscuss.FindOne(l.ctx, in.Id)
	if err != nil {
		zLog.Error("DeleteDiscuss.FindOne: get filed", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	// 获取子评论
	currentPath := fmt.Sprintf("%s%d/", discus.Path, discus.Id)
	deleteIds, count, err := l.svcCtx.TrendDiscuss.FindChildrenByPath(l.ctx, currentPath)
	if err != nil {
		zLog.Error("FindChildrenByPath failed", zap.Error(err))
		return nil, err
	}

	err = l.svcCtx.TrendDiscuss.Deletes(l.ctx, deleteIds)
	if err != nil {
		zLog.Error("Deletes failed", zap.Error(err))
		return nil, err
	}

	// 扣减当前评论的父亲评论的评论数据
	newCount := int(discus.DiscussCount) - count
	if discus.Father != 0 {
		if newCount < 0 {
			newCount = 0
		}
		err = l.svcCtx.TrendDiscuss.SetDiscuss(l.ctx, discus.Id, newCount)
		if err != nil {
			zLog.Error("SetDiscuss failed", zap.Error(err))
			return nil, err
		}
	}

	// 更新动态总评论数
	err = l.svcCtx.Trend.SetTrendReplyCount(l.ctx, discus.Id, newCount)
	if err != nil {
		zLog.Error("SetTrendReplyCount failed", zap.Error(err))
		return nil, err
	}

	return &trend.DeleteDiscussResp{
		Success: true,
	}, nil
}
