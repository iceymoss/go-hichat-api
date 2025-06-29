package comment

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDiscussLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateDiscussLogic 创建新评论
func NewCreateDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDiscussLogic {
	return &CreateDiscussLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDiscussLogic) CreateDiscuss(req *types.CreateDiscussReq) (resp *types.CreateDiscussResp, err error) {
	uid := utils.GetUser(l.ctx)
	if req.Content == "" {
		return nil, errors.New("请输入评论内容")
	}

	if req.Father == 0 { // 发表一级评论
		_, err = l.svcCtx.Trend.CreateRootDiscuss(l.ctx, &trend.CreateDiscussReq{
			TrendId:   req.TrendID,
			Father:    0,
			UserId:    req.UserID,
			Content:   req.Content,
			AtUserIds: req.AtUserIDs,
			Replyer:   uint64(uid),
		})
	} else { // 发表二级评论
		_, err = l.svcCtx.Trend.CreateChildDiscuss(l.ctx, &trend.CreateDiscussReq{
			TrendId:   req.TrendID,
			Father:    req.Father,
			UserId:    req.UserID,
			Content:   req.Content,
			AtUserIds: req.AtUserIDs,
			Replyer:   uint64(uid),
		})
	}
	if err != nil {
		zLog.Error("发表动态失败", zap.Any("req", req), zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}

	resp = &types.CreateDiscussResp{}

	return
}
