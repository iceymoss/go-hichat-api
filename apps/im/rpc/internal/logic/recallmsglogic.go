package logic

import (
	"context"
	"errors"
	"time"

	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecallMsgLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecallMsgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecallMsgLogic {
	return &RecallMsgLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RecallMsg 撤回消息：
//   - 本人撤回：受 RecallWindowSeconds 时间窗限制（0=不限）
//   - 群管理员/群主撤回：不受时间窗限制，经 social.GetMemberRole 鉴权
//
// 采用条件更新保证幂等与并发安全；返回 recalled 标记本次是否真正翻转（上游据此决定是否推送）。
func (l *RecallMsgLogic) RecallMsg(in *im.RecallMsgReq) (*im.RecallMsgResp, error) {
	if in.MsgId == "" || in.OperatorUid == "" {
		return nil, xerr.NewReqParamErr()
	}

	chatLog, err := l.svcCtx.ChatLogModel.FindOne(l.ctx, in.MsgId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, xerr.NewMsg("消息不存在")
		}
		l.Errorf("recall find msg err: %v, msgId: %s", err, in.MsgId)
		return nil, xerr.NewDBErr()
	}

	// 幂等：已撤回直接返回成功，但不再推送
	if chatLog.Status == constants.MsgStatusRecalled {
		return &im.RecallMsgResp{
			Recalled:   false,
			SendId:     chatLog.SendId,
			RecvId:     chatLog.RecvId,
			RecalledBy: chatLog.RecalledBy,
		}, nil
	}

	// 权限校验
	if err = l.checkPermission(in, chatLog); err != nil {
		return nil, err
	}

	recalled, err := l.svcCtx.ChatLogModel.UpdateRecalled(l.ctx, chatLog.ID, in.OperatorUid, time.Now().UnixNano())
	if err != nil {
		l.Errorf("recall update err: %v, msgId: %s", err, in.MsgId)
		return nil, xerr.NewDBErr()
	}

	return &im.RecallMsgResp{
		Recalled:   recalled,
		SendId:     chatLog.SendId,
		RecvId:     chatLog.RecvId,
		RecalledBy: in.OperatorUid,
	}, nil
}

// checkPermission 本人限时撤回；群管理员/群主可撤回他人消息且不限时；私聊仅本人可撤回。
func (l *RecallMsgLogic) checkPermission(in *im.RecallMsgReq, chatLog *model.ChatLog) error {
	// 本人撤回：受时间窗限制
	if in.OperatorUid == chatLog.SendId {
		window := l.svcCtx.Config.RecallWindowSeconds
		if window > 0 && time.Now().UnixNano()-chatLog.SendTime > window*int64(time.Second) {
			return xerr.NewMsg("消息发送已超过可撤回时间")
		}
		return nil
	}

	// 群聊：管理员/群主可撤回他人消息，不受时间窗限制
	if constants.ChatType(in.ChatType) == constants.GroupChatType {
		role, err := l.svcCtx.Social.GetMemberRole(l.ctx, &socialclient.GetMemberRoleReq{
			GroupId: chatLog.RecvId, // 群聊 RecvId 即 groupId
			UserId:  in.OperatorUid,
		})
		if err != nil {
			l.Errorf("recall get member role err: %v, groupId: %s, uid: %s", err, chatLog.RecvId, in.OperatorUid)
			return xerr.NewDBErr()
		}
		if role.IsMember && role.RoleLevel >= int32(constants.ManagerGroupRoleLevel) {
			return nil
		}
	}

	return xerr.NewMsg("无权限撤回该消息")
}
