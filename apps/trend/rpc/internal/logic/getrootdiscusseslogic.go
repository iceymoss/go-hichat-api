package logic

import (
	"context"
	"errors"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GetRootDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRootDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootDiscussesLogic {
	return &GetRootDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetRootDiscusses 获取一级评论（分页）
func (l *GetRootDiscussesLogic) GetRootDiscusses(in *trend.GetDiscussesReq) (*trend.DiscussesListResp, error) {
	// 验证请求参数
	if in.TrendId == 0 {
		return nil, errors.New("动态ID不能为空")
	}

	// 检查动态是否存在
	trendInfo, err := l.svcCtx.Trend.FindOne(l.ctx, in.TrendId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &trend.DiscussesListResp{}, nil
		}
		zLog.Error("GetRootDiscusses.FindOne: 检查动态存在失败", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, errors.New("动态不存在")
	}

	if trendInfo.ReplyCount > 0 {
		// 获取一级评论（使用 keyset 分页）
		discusses, err := l.svcCtx.TrendDiscuss.FindFirstByTrendId(
			l.ctx,
			in.TrendId,
			uint64(in.Pagination.LastId),
			30,
		)
		if err != nil {
			zLog.Error("GetRootDiscusses.FindByTrendWithKeyset: 获取一级评论失败", zap.Any("trendId", in.TrendId), zap.Error(err))
			return nil, errors.New("获取评论失败")
		}

		// 转换为响应格式
		respDiscusses := make([]*trend.Discuss, len(discusses))
		for i, d := range discusses {
			respDiscusses[i] = convertToReply(d)
		}

		// 设置下一页游标
		nextCursorID := uint64(0)
		nextCursorTime := int64(0)
		if len(discusses) > 0 {
			lastDiscuss := discusses[len(discusses)-1]
			nextCursorID = lastDiscuss.Id
			nextCursorTime = lastDiscuss.Createtime.Unix()
		}

		// 7. 返回响应
		return &trend.DiscussesListResp{
			Discusses: respDiscusses,
			Pagination: &trend.Pagination{
				LastId:   int32(nextCursorID),
				LastTime: nextCursorTime,
			},
		}, nil
	}

	return &trend.DiscussesListResp{}, nil

}
