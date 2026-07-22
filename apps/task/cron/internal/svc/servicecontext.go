package svc

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/config"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type RedisLocker interface {
	SetnxExCtx(ctx context.Context, key, value string, seconds int) (bool, error)
	EvalCtx(ctx context.Context, script string, keys []string, args ...any) (any, error)
}

type SocialInvitationExpirer interface {
	ExpireGroupInvitations(ctx context.Context, in *social.ExpireGroupInvitationsReq, opts ...grpc.CallOption) (*social.ExpireGroupInvitationsResp, error)
}

type ServiceContext struct {
	Config config.Config
	Social SocialInvitationExpirer
	Redis  RedisLocker
}

func NewServiceContext(c config.Config) *ServiceContext {
	var redisClient RedisLocker
	if len(c.Cache) > 0 {
		redisClient = c.Cache[0].NewRedis()
	}
	var socialRPC zrpc.Client
	if c.Cron.InvitationExpirationSpec != "" {
		secret, err := c.LoadRPCAuthSecret()
		if err != nil {
			panic(err)
		}
		rpcAuth, err := rpcauth.New(secret)
		if err != nil {
			panic(err)
		}
		socialRPC = zrpc.MustNewClient(c.SocialRpc, zrpc.WithUnaryClientInterceptor(rpcAuth.UnaryClientInterceptor()))
	} else {
		socialRPC = zrpc.MustNewClient(c.SocialRpc)
	}
	return &ServiceContext{
		Config: c,
		Social: socialclient.NewSocial(socialRPC),
		Redis:  redisClient,
	}
}
