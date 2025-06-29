package comment

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkDiscussReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMarkDiscussReadLogic 标记评论为已读
func NewMarkDiscussReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkDiscussReadLogic {
	return &MarkDiscussReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkDiscussReadLogic) MarkDiscussRead(req *types.MarkDiscussRequest) (resp *types.MarkDiscussResponse, err error) {
	uid := utils.GetUser(l.ctx)
	_, err = l.svcCtx.Trend.MarkDiscussRead(l.ctx, &trend.MarkDiscussRequest{
		UserId: uint32(uid),
		DisIds: req.DisIDs,
	})
	if err != nil {
		zLog.Error("标记动态评论已读失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	resp = &types.MarkDiscussResponse{}

	return
}
