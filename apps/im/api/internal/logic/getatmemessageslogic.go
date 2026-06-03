package logic

import (
	"context"
	"encoding/base64"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetAtMeMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetAtMeMessagesLogic 获取会话中@我且未读的消息列表(用于有人@我快速跳转)
func NewGetAtMeMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAtMeMessagesLogic {
	return &GetAtMeMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAtMeMessagesLogic) GetAtMeMessages(req *types.GetAtMeMessagesReq) (resp *types.GetAtMeMessagesResp, err error) {
	uid := l.ctx.Value(Identify).(string)

	res, err := l.svcCtx.IM.GetAtMeMessages(l.ctx, &im.GetAtMeMessagesReq{
		ConversationId: req.ConversationId,
		UserId:         uid,
		Count:          req.Count,
	})
	if err != nil {
		zLog.Error("get at-me messages failed", zap.Error(err))
		return nil, err
	}

	list := make([]types.ChatLog, 0, len(res.List))
	for _, v := range res.List {
		var readRec string
		if len(v.ReadRecords) > 0 {
			readRec = base64.StdEncoding.EncodeToString(v.ReadRecords)
		}
		list = append(list, types.ChatLog{
			Id:             v.Id,
			ConversationId: v.ConversationId,
			SendId:         v.SendId,
			RecvId:         v.RecvId,
			MsgType:        v.MsgType,
			MsgContent:     v.MsgContent,
			ChatType:       v.ChatType,
			SendTime:       v.SendTime,
			ReadRecords:    readRec,
			Quote:          v.Quote,
			Status:         v.Status,
			RecalledBy:     v.RecalledBy,
			AtUsers:        v.AtUsers,
			AtAll:          v.AtAll,
		})
	}

	return &types.GetAtMeMessagesResp{List: list}, nil
}
