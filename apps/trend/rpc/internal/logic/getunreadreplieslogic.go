package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUnreadRepliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadRepliesLogic {
	return &GetUnreadRepliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUnreadReplies 获取未读回复
func (l *GetUnreadRepliesLogic) GetUnreadReplies(in *trend.GetUnreadRepliesReq) (*trend.RepliesListResp, error) {
	if in.UserId == 0 {
		return nil, errors.New("请登录")
	}

	unReadList, err := l.svcCtx.TrendDiscuss.FindUnreadByUser(l.ctx, int(in.UserId), int(in.Pagination.LastId))
	if err != nil {
		return nil, errors.New("未获取到未读消息")
	}

	list := make([]*trend.Discuss, 0, len(unReadList))
	idList := make([]uint64, 0, len(unReadList))
	for _, v := range unReadList {
		list = append(list, convertToReply(v))
		idList = append(idList, v.Id)
	}

	// 更新未读状态,标记为已读
	err = l.svcCtx.TrendDiscuss.MarkReadById(l.ctx, idList)
	if err != nil {
		zLog.Error("GetUnreadReplies.MarkReadById: 标记为已读失败", zap.Any("idList", idList), zap.Error(err))
		return nil, err
	}

	return &trend.RepliesListResp{
		Replies: list,
		Total:   uint64(len(list)),
	}, nil
}
