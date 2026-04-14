package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatLogLogic {
	return &GetChatLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetChatLog 获取会话记录
// 依据传递的数据时间点获取
func (l *GetChatLogLogic) GetChatLog(in *im.GetChatLogReq) (*im.GetChatLogResp, error) {
	// msgId 作为游标：查该消息的 sendTime，然后查更早的 count 条
	if in.MsgId != "" {
		chatlog, err := l.svcCtx.ChatLogModel.FindOne(l.ctx, in.MsgId)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog by msgId err %v, req %v", err, in.MsgId)
		}

		// 用该消息的 sendTime - 1 作为 endSendTime，查更早的消息
		endTime := chatlog.SendTime - 1
		data, err := l.svcCtx.ChatLogModel.ListBySendTime(l.ctx, in.ConversationId, in.StartSendTime, endTime, in.Count)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog list by cursor err %v", err)
		}

		res := make([]*im.ChatLog, 0, len(data))
		for _, datum := range data {
			res = append(res, &im.ChatLog{
				Id:             datum.ID.Hex(),
				ConversationId: datum.ConversationId,
				SendId:         datum.SendId,
				RecvId:         datum.RecvId,
				MsgType:        int32(datum.MsgType),
				MsgContent:     datum.MsgContent,
				ChatType:       int32(datum.ChatType),
				SendTime:       datum.SendTime,
				ReadRecords:    datum.ReadRecords,
			})
		}

		return &im.GetChatLogResp{List: res}, nil
	}

	// 无游标：根据会话和时间范围查询
	data, err := l.svcCtx.ChatLogModel.ListBySendTime(l.ctx, in.ConversationId, in.StartSendTime, in.EndSendTime, in.Count)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog list by SendTime err %v, req %v", err, in)
	}

	res := make([]*im.ChatLog, 0, len(data))
	for _, datum := range data {
		res = append(res, &im.ChatLog{
			Id:             datum.ID.Hex(),
			ConversationId: datum.ConversationId,
			SendId:         datum.SendId,
			RecvId:         datum.RecvId,
			MsgType:        int32(datum.MsgType),
			MsgContent:     datum.MsgContent,
			ChatType:       int32(datum.ChatType),
			SendTime:       datum.SendTime,
			ReadRecords:    datum.ReadRecords,
		})
	}

	return &im.GetChatLogResp{
		List: res,
	}, nil
}
