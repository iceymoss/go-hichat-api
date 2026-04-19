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

type GetChatLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetChatLogLogic 根据用户获取聊天记录
func NewGetChatLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatLogLogic {
	return &GetChatLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatLogLogic) GetChatLog(req *types.ChatLogReq) (resp *types.ChatLogResp, err error) {
	// todo: add your logic here and delete this line
	res, err := l.svcCtx.IM.GetChatLog(l.ctx, &im.GetChatLogReq{
		ConversationId: req.ConversationId,
		StartSendTime:  req.StartSendTime,
		EndSendTime:    req.EndSendTime,
		Count:          req.Count,
		MsgId:          req.MsgId,
	})
	if err != nil {
		zLog.Error("get chatlog failed", zap.Error(err))
		return nil, err
	}

	list := make([]types.ChatLog, 0, len(res.List))

	for _, v := range res.List {
		// ReadRecords 是 bitmap 字节，base64 编码成字符串透传给前端
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
		})
	}

	resp = &types.ChatLogResp{List: list}

	return
}
