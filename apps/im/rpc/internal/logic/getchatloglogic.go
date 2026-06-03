package logic

import (
	"context"

	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

func toChatLogPb(c *model.ChatLog) *im.ChatLog {
	pb := &im.ChatLog{
		Id:             c.ID.Hex(),
		ConversationId: c.ConversationId,
		SendId:         c.SendId,
		RecvId:         c.RecvId,
		MsgType:        int32(c.MsgType),
		MsgContent:     c.MsgContent,
		ChatType:       int32(c.ChatType),
		SendTime:       c.SendTime,
		ReadRecords:    c.ReadRecords,
		ReadTimes:      c.ReadTimes,
		Quote:          c.Quote,
		Status:         int32(c.Status),
		RecalledBy:     c.RecalledBy,
		AtUsers:        c.AtUsers,
		AtAll:          c.AtAll,
	}
	// 已撤回消息：原文保留在库中作审计，但绝不下发；引用预览同理交由前端按 status 降级。
	if c.Status == constants.MsgStatusRecalled {
		pb.MsgContent = ""
		pb.Quote = ""
	}
	return pb
}

func toChatLogPbList(list []*model.ChatLog) []*im.ChatLog {
	res := make([]*im.ChatLog, 0, len(list))
	for _, c := range list {
		res = append(res, toChatLogPb(c))
	}
	return res
}

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
	// msgId 作为游标
	if in.MsgId != "" {
		target, err := l.svcCtx.ChatLogModel.FindOne(l.ctx, in.MsgId)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog by msgId err %v, req %v", err, in.MsgId)
		}

		switch in.Direction {
		case "newer":
			// 向下增量：查严格晚于目标的 count 条（按时间升序）
			data, err := l.svcCtx.ChatLogModel.ListAfter(l.ctx, in.ConversationId, target.SendTime, in.Count)
			if err != nil {
				return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog newer err %v", err)
			}
			return &im.GetChatLogResp{List: toChatLogPbList(data)}, nil
		case "around":
			// 跳转初始窗口：目标消息 + 其前 count-1 条（含目标，按时间降序）
			data, err := l.svcCtx.ChatLogModel.ListBySendTime(l.ctx, in.ConversationId, 0, target.SendTime, in.Count)
			if err != nil {
				return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog around err %v", err)
			}
			return &im.GetChatLogResp{List: toChatLogPbList(data)}, nil
		}

		// 仅传 MsgId（Count=0 且无时间范围）→ 精确单条查询（供已读详情等点查场景）
		if in.Count == 0 && in.StartSendTime == 0 && in.EndSendTime == 0 {
			return &im.GetChatLogResp{List: []*im.ChatLog{toChatLogPb(target)}}, nil
		}

		// 默认 older：查更早的 count 条
		data, err := l.svcCtx.ChatLogModel.ListBySendTime(l.ctx, in.ConversationId, in.StartSendTime, target.SendTime-1, in.Count)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog list by cursor err %v", err)
		}
		return &im.GetChatLogResp{List: toChatLogPbList(data)}, nil
	}

	// 无游标：根据会话和时间范围查询
	data, err := l.svcCtx.ChatLogModel.ListBySendTime(l.ctx, in.ConversationId, in.StartSendTime, in.EndSendTime, in.Count)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find chatLog list by SendTime err %v, req %v", err, in)
	}
	return &im.GetChatLogResp{List: toChatLogPbList(data)}, nil
}
