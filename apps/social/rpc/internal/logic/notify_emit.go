package logic

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

// 公共通知 notifyType 常量（与前端 ws.on('notify') 分发的 type 一一对应）。
const (
	NotifyFriendApply  = "friend.apply"  // 收到好友申请 -> 通知被申请人
	NotifyFriendAccept = "friend.accept" // 好友申请通过 -> 通知申请人
	NotifyFriendReject = "friend.reject" // 好友申请拒绝 -> 通知申请人
	NotifyGroupApply   = "group.apply"   // 收到入群申请 -> 通知群主/管理员
	NotifyGroupAccept  = "group.accept"  // 入群申请通过 -> 通知申请人
	NotifyGroupReject  = "group.reject"  // 入群申请拒绝 -> 通知申请人

	NotifyGroupRemoved          = "group.removed"           // 被移出群 -> 通知被移出者
	NotifyGroupAdminSet         = "group.admin.set"         // 被设为管理员 -> 通知本人
	NotifyGroupAdminUnset       = "group.admin.unset"       // 被取消管理员 -> 通知本人
	NotifyGroupOwnerTransferred = "group.owner.transferred" // 成为新群主 -> 通知新群主
)

// emitCommonNotify 把一条公共通知直接 Push 到 Kafka（参考撤回范式）。
// best-effort：失败仅打日志、不阻断业务（通知非关键路径，离线/丢失由落库 + 上线拉列表兜底）。
// 防御：空接收者 / 自己通知自己直接跳过。
func emitCommonNotify(ctx context.Context, svcCtx *svc.ServiceContext, n *mq.CommonNotify) {
	if svcCtx.CommonNotifyClient == nil {
		return
	}
	if n.ReceiverId == "" || n.ReceiverId == n.ActorId {
		return
	}
	if n.CreateTime == 0 {
		n.CreateTime = time.Now().Unix()
	}
	if err := svcCtx.CommonNotifyClient.Push(n); err != nil {
		logx.WithContext(ctx).Errorf("emitCommonNotify push failed: type=%s receiver=%s biz=%s err=%v",
			n.NotifyType, n.ReceiverId, n.BizId, err)
	}
}

// notifyGroupAdmins 把入群申请通知扇出给该群的群主 + 全部管理员（每人一条单接收者消息）。
func notifyGroupAdmins(ctx context.Context, svcCtx *svc.ServiceContext, groupId, applicantId, bizId, content string) {
	members, err := svcCtx.GroupMembersModel.ListByGroupId(ctx, groupId)
	if err != nil {
		logx.WithContext(ctx).Errorf("notifyGroupAdmins list members failed: group=%s err=%v", groupId, err)
		return
	}
	for _, m := range members {
		// 仅群主 + 管理员（role_level >= ManagerGroupRoleLevel）
		if constants.GroupRoleLevel(m.RoleLevel) < constants.ManagerGroupRoleLevel {
			continue
		}
		emitCommonNotify(ctx, svcCtx, &mq.CommonNotify{
			NotifyType: NotifyGroupApply,
			ReceiverId: m.UserId,
			ActorId:    applicantId,
			BizId:      bizId,
			GroupId:    groupId,
			Content:    content,
		})
	}
}

// normalizeGroupMemberUid 把空的 inviter_uid/operator_uid 归一为 "0"（写入 INT 列为 0）。
// 原因：列是 INT，空串写入报 Error 1366；而模型字段是 string、读取无法把 NULL 扫进 string，
// 所以用 "0" 哨兵（无邀请人/操作人）既能写也能读，与全仓把这两列当非空 string 的约定一致。
func normalizeGroupMemberUid(gm *socialmodels.GroupMembers) {
	if gm.InviterUid == "" {
		gm.InviterUid = "0"
	}
	if gm.OperatorUid == "" {
		gm.OperatorUid = "0"
	}
}
