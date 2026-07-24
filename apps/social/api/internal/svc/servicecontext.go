package svc

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/imclient"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/presence"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type UserLookup interface {
	GetUserById(ctx context.Context, in *user.GetUserByIdRequest, opts ...grpc.CallOption) (*user.GetUserByIdResponse, error)
	FindUser(ctx context.Context, in *user.FindUserReq, opts ...grpc.CallOption) (*user.FindUserResp, error)
}

// ServiceContext rpc配置，需要调用的模块都需要在这里配置
type ServiceContext struct {
	Config   config.Config
	Social   socialclient.Social
	User     UserLookup
	Im       imclient.Im
	Presence *presence.Store
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		Social:   socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:     userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		Im:       imclient.NewIm(zrpc.MustNewClient(c.ImRpc)),
		Presence: presence.New(db.GetRedisConn()),
	}
}
