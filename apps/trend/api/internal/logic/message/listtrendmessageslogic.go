package message

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type ListTrendMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListTrendMessagesLogic 获取动态消息列表
func NewListTrendMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrendMessagesLogic {
	return &ListTrendMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTrendMessagesLogic) ListTrendMessages(req *types.ListTrendMessagesReq) (resp *types.ListTrendMessagesResp, err error) {
	uid := utils.GetUser(l.ctx)
	rpcResp, err := l.svcCtx.Trend.ListTrendMessages(l.ctx, &trend.ListTrendMessagesReq{
		UserId: uint64(uid),
		LastId: int32(req.LastId),
		Limit:  int32(req.Limit),
	})
	if err != nil {
		zLog.Error("ListTrendMessages: rpc failed", zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}

	// 批量取触发者用户信息
	actorIds := make([]string, 0, len(rpcResp.Messages))
	for _, m := range rpcResp.Messages {
		actorIds = append(actorIds, strconv.Itoa(int(m.ActorId)))
	}

	userBind := make(map[string]*types.User, len(actorIds))
	if len(actorIds) > 0 {
		userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{Ids: actorIds})
		if err != nil {
			zLog.Error("ListTrendMessages: find user failed", zap.Any("ids", actorIds), zap.Error(err))
			return nil, err
		}
		for _, v := range userInfo.User {
			userBind[v.Id] = &types.User{
				Id:       v.Id,
				Nickname: v.Nickname,
				Sex:      int(v.Sex),
				Avatar:   v.Avatar,
			}
		}
	}

	list := make([]*types.TrendMessageItem, 0, len(rpcResp.Messages))
	for _, m := range rpcResp.Messages {
		list = append(list, &types.TrendMessageItem{
			Id:              m.Id,
			Type:            int(m.Type),
			TrendId:         m.TrendId,
			CommentId:       m.CommentId,
			ParentCommentId: m.ParentCommentId,
			Content:         m.Content,
			IsRead:          m.IsRead,
			CreateTime:      m.CreateTime,
			Actor:           userBind[strconv.Itoa(int(m.ActorId))],
		})
	}

	resp = &types.ListTrendMessagesResp{
		List:   list,
		LastId: int(rpcResp.LastId),
	}
	return
}
