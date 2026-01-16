package logic

import (
	"context"

	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLogic {
	return &GroupInviteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupInviteLogic) GroupInvite(in *social.GroupInviteReq) (*social.GroupInviteResp, error) {
	// 说明：邀请入群在业务上属于 GroupPutin 的一个分支（joinSource=邀请）。
	// 这里保留 GroupInvite 只是为了 API/前端兼容，内部统一走 GroupPutin，避免两套逻辑不一致。

	// inviter 必须是群成员（否则不允许“替别人申请入群”）
	_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("inviter not in group"), "inviter not in group")
	}

	putin := NewGroupPutinLogic(l.ctx, l.svcCtx)
	now := time.Now().Unix()

	for _, friendId := range in.FriendIds {
		_, err := putin.GroupPutin(&social.GroupPutinReq{
			GroupId:    in.GroupId,
			ReqId:      friendId, // 被邀请者
			ReqMsg:     "invited",
			ReqTime:    now,
			JoinSource: int32(constants.InviteGroupJoinSource),
			InviterUid: in.UserId, // 邀请者
		})
		if err != nil {
			// 不阻塞其它邀请，记录日志即可
			l.Logger.Errorf("GroupInvite->GroupPutin failed: groupId=%s inviter=%s friend=%s err=%v", in.GroupId, in.UserId, friendId, err)
		}
	}

	return &social.GroupInviteResp{}, nil
}
