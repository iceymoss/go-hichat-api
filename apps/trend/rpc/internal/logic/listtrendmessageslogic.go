package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type ListTrendMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTrendMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrendMessagesLogic {
	return &ListTrendMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListTrendMessages 获取动态消息列表（id 倒序分页）
func (l *ListTrendMessagesLogic) ListTrendMessages(in *trend.ListTrendMessagesReq) (*trend.ListTrendMessagesResp, error) {
	list, err := l.svcCtx.TrendMessage.ListByReceiver(l.ctx, in.UserId, int(in.LastId), int(in.Limit))
	if err != nil {
		zLog.Error("ListTrendMessages: list failed", zap.Uint64("userId", in.UserId), zap.Error(err))
		return nil, err
	}

	resp := &trend.ListTrendMessagesResp{
		Messages: make([]*trend.TrendMessageInfo, 0, len(list)),
	}
	for _, m := range list {
		resp.Messages = append(resp.Messages, trendMessageToInfo(m))
		resp.LastId = int32(m.Id)
	}
	return resp, nil
}

// trendMessageToInfo model -> proto
func trendMessageToInfo(m *models.TrendMessage) *trend.TrendMessageInfo {
	return &trend.TrendMessageInfo{
		Id:              m.Id,
		ReceiverId:      m.ReceiverId,
		ActorId:         m.ActorId,
		Type:            int32(m.Type),
		TrendId:         m.TrendId,
		CommentId:       m.CommentId,
		ParentCommentId: m.ParentCommentId,
		Content:         m.Content,
		IsRead:          m.IsRead == 1,
		CreateTime:      m.CreateTime.Unix(),
	}
}
