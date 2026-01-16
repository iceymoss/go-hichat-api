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

type GroupPutInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGroupPutInLogic 申请进群
func NewGroupPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInLogic {
	return &GroupPutInLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupPutInLogic) GroupPutIn(req *types.GroupPutInReq) (resp *types.GroupPutInResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	
	// 如果传了 token，走 token 入群流程（内部会调用 GroupPutin）
	if req.Token != "" {
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
					zLog.Error("GroupPutIn.SetUpUserConversation: best-effort failed", zap.Error(err))
				}
			}(uid, recvId)
		}
		
		return &types.GroupPutInResp{
			GroupId: int(rpcResp.GroupId),
			IsPass:  int(rpcResp.IsPass),
		}, nil
	}
	
	// 普通申请/邀请入群流程
	// 如果传了 reqId（邀请入群场景），使用 reqId；否则使用当前用户ID（普通申请场景）
	reqId := req.ReqId
	if reqId == "" {
		reqId = uid // 普通申请：使用当前用户ID
	}
	
	res, err := l.svcCtx.Social.GroupPutin(l.ctx, &social.GroupPutinReq{
		GroupId:    req.GroupId,           // 群id
		ReqId:      reqId,                  // 请求者/被邀请者ID
		ReqMsg:     req.ReqMsg,            // 请求消息
		ReqTime:    time.Now().Unix(),     //请求时间
		JoinSource: int32(req.JoinSource), //请求来源
		InviterUid: req.InviterUid,        //邀请人
	})
	if err != nil {
		return nil, err
	}

	// 如果成功加入群聊后，为其用户创建该群的聊天会话
	if res.IsPass == 1 {
		recvId := strconv.Itoa(int(res.GroupId))
		// Best-effort: do NOT block or fail join if im-rpc is down.
		// 使用 reqId（被邀请者）创建会话，而不是当前用户
		go func(sendId, recvId string) {
			ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
			defer cancel()
			if _, err := l.svcCtx.Im.SetUpUserConversation(ctx, &im.SetUpUserConversationReq{
				SendId:   sendId,
				RecvId:   recvId,
				ChatType: int32(constants.GroupChatType),
			}); err != nil {
				zLog.Error("GroupPutIn.SetUpUserConversation: best-effort failed", zap.Error(err))
			}
		}(reqId, recvId)
	}

	resp = &types.GroupPutInResp{
		GroupId: int(res.GroupId),
		IsPass:  int(res.IsPass),
	}
	return
}
