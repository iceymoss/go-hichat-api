package group

import (
	"context"
	"go.uber.org/zap"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
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
	uid := ctxdata.GetUId(l.ctx)
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

	// 非同意则不需要建立群会话
	if req.HandleResult != int32(constants.PassHandlerResult) {
		return &types.GroupPutInHandleResp{}, nil
	}

	// 为当前申请的用户添加群聊会话
	// Best-effort: do NOT block or fail approval if im-rpc is down.
	go func(sendId, recvId string) {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		defer cancel()
		if _, err := l.svcCtx.Im.SetUpUserConversation(ctx, &im.SetUpUserConversationReq{
			SendId:   sendId,
			RecvId:   recvId,
			ChatType: int32(constants.GroupChatType),
		}); err != nil {
			zLog.Error("GroupPutInHandle.SetUpUserConversation: best-effort failed", zap.Error(err))
		}
	}(res.ReqId, res.GroupId)

	return &types.GroupPutInHandleResp{}, nil
}
