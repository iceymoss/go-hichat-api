package config

import (
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	Cache cache.CacheConf

	ImRpc zrpc.RpcClientConf

	UserRpc       zrpc.RpcClientConf
	RpcAuthSecret string `json:",optional"`

	// RelationChangeTransfer 关系变更事件 Kafka 生产端（relay 投递用）
	RelationChangeTransfer struct {
		Addrs []string
		Topic string
	}

	// CommonNotifyTransfer 公共通知事件 Kafka 生产端（好友/群申请等实时通知）
	CommonNotifyTransfer struct {
		Addrs []string
		Topic string
	}

	SocialRequestNotification struct {
		Addrs []string `json:",optional"`
		Topic string   `json:",optional"`
	}
}

const rpcAuthSecretEnv = "HICHAT_SOCIAL_RPC_AUTH_SECRET"

func LoadRPCAuthSecret(configured string) (string, error) {
	secret := configured
	if value, ok := os.LookupEnv(rpcAuthSecretEnv); ok && value != "" {
		secret = value
	}
	if len(secret) < 32 {
		return "", fmt.Errorf("social RPC auth secret must be at least 32 bytes")
	}
	return secret, nil
}

func OptionalRPCAuthSecret(configured string) string {
	secret := configured
	if value, ok := os.LookupEnv(rpcAuthSecretEnv); ok && value != "" {
		secret = value
	}
	if len(secret) < 32 {
		return ""
	}
	return secret
}
