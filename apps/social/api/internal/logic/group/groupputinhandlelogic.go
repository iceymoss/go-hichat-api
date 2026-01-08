package group

import (
	"context"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInHandleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGroupPutInHandleLogic 申请进群处理
func NewGroupPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInHandleLogic {
	return &GroupPutInHandleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupPutInHandleLogic) GroupPutInHandle(req *types.GroupPutInHandleReq) (resp *types.GroupPutInHandleResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.Social.GroupPutInHandle(l.ctx, &social.GroupPutInHandleReq{
		GroupReqId:   req.GroupReqId,
		GroupId:      req.GroupId,
		HandleUid:    uid,
		HandleResult: req.HandleResult,
	})
	if err != nil {
		zLog.Error("GroupPutInHandle.GroupPutInHandle: groupPutInHandle failed", zap.Error(err))
		return nil, err
	}

	// 为当前申请的用户添加群聊会话
	_, err = l.svcCtx.Im.SetUpUserConversation(l.ctx, &im.SetUpUserConversationReq{
		SendId:   uid,
		RecvId:   res.GroupId,
		ChatType: int32(constants.GroupChatType),
	})
	if err != nil {
		zLog.Error("GroupPutInHandle.SetUpUserConversation: create conversation failed", zap.Error(err))
		return nil, err
	}

	return
}
