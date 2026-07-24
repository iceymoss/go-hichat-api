package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"
)

const (
	groupInvitationExpirationLockKey = "task:cron:social:group-invitation-expiration"
	compareDeleteScript              = `if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`
)

type GroupInvitationExpirationTask struct {
	svc *svc.ServiceContext
}

func NewGroupInvitationExpirationTask(svcCtx *svc.ServiceContext) *GroupInvitationExpirationTask {
	return &GroupInvitationExpirationTask{svc: svcCtx}
}

func (t *GroupInvitationExpirationTask) GetName() string { return "group_invitation_expiration" }

func (t *GroupInvitationExpirationTask) GetSpec() string {
	return t.svc.Config.Cron.InvitationExpirationSpec
}

func (t *GroupInvitationExpirationTask) GetDescription() string {
	return "Expire pending Social group invitations in bounded batches"
}

func (t *GroupInvitationExpirationTask) GetTimeout() time.Duration {
	return t.svc.Config.TaskTimeout()
}

func (t *GroupInvitationExpirationTask) Execute(ctx context.Context) (err error) {
	if t.svc.Redis == nil {
		return fmt.Errorf("group invitation expiration Redis lock is not configured")
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("create group invitation expiration lock token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	acquired, err := t.svc.Redis.SetnxExCtx(ctx, groupInvitationExpirationLockKey, token, t.svc.Config.Cron.LockTTLSeconds)
	if err != nil {
		return fmt.Errorf("acquire group invitation expiration lock: %w", err)
	}
	if !acquired {
		return nil
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_, unlockErr := t.svc.Redis.EvalCtx(unlockCtx, compareDeleteScript, []string{groupInvitationExpirationLockKey}, token)
		if unlockErr != nil && err == nil {
			err = fmt.Errorf("release group invitation expiration lock: %w", unlockErr)
		}
	}()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := t.svc.Social.ExpireGroupInvitations(rpcauth.WithTask(ctx), &social.ExpireGroupInvitationsReq{BatchSize: int32(t.svc.Config.Cron.BatchSize)})
		if err != nil {
			return fmt.Errorf("expire group invitations: %w", err)
		}
		if resp == nil {
			return fmt.Errorf("expire group invitations returned a nil response")
		}
		if !resp.HasMore {
			return nil
		}
	}
}
