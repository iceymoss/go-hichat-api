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
		// ChatLog.SendTime 历史上由 Insert 默认写成 UnixMilli，其它路径可能是 UnixNano，
		// 统一归一到毫秒再比较，避免单位不一致导致永远判超时。SendTime<=0（未知）则不限制。
		sentMs := normalizeUnixMillis(chatLog.SendTime)
		if window > 0 && sentMs > 0 && time.Now().UnixMilli()-sentMs > window*1000 {
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

// normalizeUnixMillis 把可能为秒 / 毫秒 / 微秒 / 纳秒的 unix 时间戳统一归一到毫秒。
// 历史数据里 ChatLog.SendTime 单位不一致（Insert 默认写毫秒，旧逻辑写过纳秒），
// 按数量级归一可避免单位不一致导致的时间窗误判。ts<=0 视为未知，返回 0。
func normalizeUnixMillis(ts int64) int64 {
	if ts <= 0 {
		return 0
	}
	switch {
	case ts >= 1e18: // 纳秒
		return ts / 1e6
	case ts >= 1e15: // 微秒
		return ts / 1e3
	case ts >= 1e12: // 毫秒
		return ts
	default: // 秒
		return ts * 1000
	}
}
