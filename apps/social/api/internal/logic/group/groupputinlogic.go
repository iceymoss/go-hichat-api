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

func (l *GroupPutInLogic) GroupPutIn(req *types.GroupPutInRep) (resp *types.GroupPutInResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.Social.GroupPutin(l.ctx, &social.GroupPutinReq{
		GroupId:    req.GroupId,           // 群id
		ReqId:      uid,                   // 请求者
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
		_, err = l.svcCtx.Im.SetUpUserConversation(l.ctx, &im.SetUpUserConversationReq{
			SendId:   uid,
			RecvId:   strconv.Itoa(int(res.GroupId)),
			ChatType: int32(constants.GroupChatType),
		})
		if err != nil {
			zLog.Error("GroupPutIn.SetUpUserConversation: create conversation failed", zap.Error(err))
			return nil, err
		}
	}

	resp = &types.GroupPutInResp{
		GroupId: int(res.GroupId),
	}
	return
}
