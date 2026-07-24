package logic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLinkCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupInviteLinkCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLinkCreateLogic {
	return &GroupInviteLinkCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 邀请链接/二维码入群
func (l *GroupInviteLinkCreateLogic) GroupInviteLinkCreate(in *social.GroupInviteLinkCreateReq) (*social.GroupInviteLinkCreateResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	// 1) 校验群存在
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found %v", in.GroupId)
	}
	if group.Status == 1 {
		return nil, errors.Wrapf(xerr.NewMsg("group disbanded"), "group disbanded")
	}

	// 2) 权限：至少管理员/群主才能创建邀请链接（后续可扩展为群设置开关）
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, actor, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}
	if member.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only admin/owner can create invite link")
	}

	groupIdInt, err := strconv.Atoi(in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}
	uidInt, err := strconv.ParseUint(actor, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid userId"), "invalid userId")
	}

	maxUses := in.MaxUses
	if maxUses < 0 {
		maxUses = 0
	}

	var expireAt *time.Time
	if in.ExpireSeconds > 0 {
		t := time.Now().Add(time.Duration(in.ExpireSeconds) * time.Second)
		expireAt = &t
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)

	// 3) 生成唯一 token（重试）
	var token string
	for i := 0; i < 5; i++ {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return nil, errors.Wrapf(xerr.NewMsg("token generate failed"), "rand read failed")
		}
		token = base64.RawURLEncoding.EncodeToString(b) // ~43 chars

		res := mysqlConn.Table("group_invite_links").Create(map[string]any{
			"group_id":   groupIdInt,
			"token":      token,
			"created_by": uidInt,
			"expire_at":  expireAt,
			"max_uses":   maxUses,
			"used_count": 0,
			"revoked":    0,
			"revoked_at": nil,
			"created_at": time.Now(),
		})
		if res.Error == nil && res.RowsAffected > 0 {
			break
		}
		// duplicate token -> retry, otherwise return error
		if res.Error != nil {
			// 简单判断重复键（不同驱动错误信息不同），失败则重试几次
			if strings.Contains(res.Error.Error(), "Duplicate") || strings.Contains(res.Error.Error(), "duplicate") {
				continue
			}
			return nil, errors.Wrapf(xerr.NewDBErr(), "create invite link err %v", res.Error)
		}
	}
	if token == "" {
		return nil, errors.Wrapf(xerr.NewMsg("token generate failed"), "token empty")
	}

	now := time.Now().Unix()
	exp := int64(0)
	if expireAt != nil {
		exp = expireAt.Unix()
	}

	return &social.GroupInviteLinkCreateResp{
		Link: &social.GroupInviteLink{
			Token:     token,
			GroupId:   in.GroupId,
			CreatedBy: actor,
			CreatedAt: now,
			ExpireAt:  exp,
			MaxUses:   maxUses,
			UsedCount: 0,
			Revoked:   false,
			RevokedAt: 0,
		},
	}, nil
}
