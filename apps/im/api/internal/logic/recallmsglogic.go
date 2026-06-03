package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type RecallMsgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 撤回消息
func NewRecallMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecallMsgLogic {
	return &RecallMsgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RecallMsg 编排：调 im-rpc 完成校验+撤回（业务在 rpc），成功且本次真正翻转时，
// 向通用 MsgChatTransfer 链路投递一条 ContentRecall 控制帧做推送（状态变更与推送副作用分离）。
func (l *RecallMsgLogic) RecallMsg(req *types.RecallMsgReq) (resp *types.RecallMsgResp, err error) {
	uid := l.ctx.Value(Identify).(string)

	rpcResp, err := l.svcCtx.IM.RecallMsg(l.ctx, &im.RecallMsgReq{
		ConversationId: req.ConversationId,
		MsgId:          req.MsgId,
		OperatorUid:    uid,
		ChatType:       req.ChatType,
	})
	if err != nil {
		return nil, err
	}

	// 幂等：已撤回（recalled=false）不再重复推送
	if rpcResp.Recalled {
		if perr := l.publishRecall(req, rpcResp); perr != nil {
			// 推送失败不回滚撤回（DB 已撤回），仅记录；其他端可通过历史拉取补偿
			zLog.Error("publish recall event failed", zap.Error(perr), zap.String("msgId", req.MsgId))
		}
	}

	return &types.RecallMsgResp{}, nil
}

// publishRecall 复用 MsgChatTransfer 通用 topic，以 ContentRecall 控制帧下发；
// 路由沿用原消息的 SendId/RecvId/ChatType，消费端据 MsgType 分流（跳过落库，仅扇出）。
func (l *RecallMsgLogic) publishRecall(req *types.RecallMsgReq, rpcResp *im.RecallMsgResp) error {
	return l.svcCtx.MsgChatTransferClient.Push(&mq.MsgChatTransfer{
		ChatType:       constants.ChatType(req.ChatType),
		ConversationId: req.ConversationId,
		SendId:         rpcResp.SendId,
		RecvId:         rpcResp.RecvId,
		MsgId:          req.MsgId,
		MsgType:        constants.ContentRecall,
		ContentType:    constants.ContentRecall,
		RecalledBy:     rpcResp.RecalledBy,
	})
}
