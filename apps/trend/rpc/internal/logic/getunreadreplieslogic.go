package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
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

	var lastId uint64
	if len(list) > 0 {
		lastId = list[len(list)-1].Id
	}

	return &trend.RepliesListResp{
		Replies: list,
		Total:   uint64(len(list)),
		Pagination: &trend.Pagination{
			LastId: int32(lastId),
		},
	}, nil
}
