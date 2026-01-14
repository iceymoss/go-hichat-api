package group

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GroupJoinByTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupJoinByTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupJoinByTokenLogic {
	return &GroupJoinByTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupJoinByTokenLogic) GroupJoinByToken(req *types.GroupJoinByTokenReq) (resp *types.GroupJoinByTokenResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GroupJoinByToken(l.ctx, &social.GroupJoinByTokenReq{
		UserId: uid,
		Token:  req.Token,
		ReqMsg: req.ReqMsg,
	})
	if err != nil {
		return nil, err
	}

	// direct join: setup conversation
	if rpcResp.IsPass == 1 {
		recvId := strconv.Itoa(int(rpcResp.GroupId))
		go func(sendId, recvId string) {
			ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
			defer cancel()
			if _, err := l.svcCtx.Im.SetUpUserConversation(ctx, &im.SetUpUserConversationReq{
				SendId:   sendId,
				RecvId:   recvId,
				ChatType: int32(constants.GroupChatType),
			}); err != nil {
				zLog.Error("GroupJoinByToken.SetUpUserConversation: best-effort failed", zap.Error(err))
			}
		}(uid, recvId)
	}

	return &types.GroupJoinByTokenResp{
		GroupId: rpcResp.GroupId,
		IsPass:  rpcResp.IsPass,
	}, nil
}
