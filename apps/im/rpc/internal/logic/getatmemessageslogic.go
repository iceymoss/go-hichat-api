package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/bitmap"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAtMeMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAtMeMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAtMeMessagesLogic {
	return &GetAtMeMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAtMeMessages 获取会话中 @我 且我尚未读的消息列表（用于"有人@我"快速跳转）。
// 先按 sendTime 降序取最近若干条 @我 消息，再按 ReadRecords bitmap 过滤掉我已读的，
// 最后反转为升序（最早的 @ 在前），方便前端从上到下逐条跳转。
func (l *GetAtMeMessagesLogic) GetAtMeMessages(in *im.GetAtMeMessagesReq) (*im.GetAtMeMessagesResp, error) {
	if in.ConversationId == "" || in.UserId == "" {
		return &im.GetAtMeMessagesResp{List: []*im.ChatLog{}}, nil
	}

	list, err := l.svcCtx.ChatLogModel.ListAtMe(l.ctx, in.ConversationId, in.UserId, in.Count)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list at-me messages err %v, req %v", err, in)
	}

	// 降序结果里过滤出"我未读"的；倒序遍历同时完成升序输出。
	res := make([]*im.ChatLog, 0, len(list))
	for i := len(list) - 1; i >= 0; i-- {
		c := list[i]
		if bitmap.Load(c.ReadRecords).IsSet(in.UserId) {
			continue
		}
		res = append(res, toChatLogPb(c))
	}

	return &im.GetAtMeMessagesResp{List: res}, nil
}
