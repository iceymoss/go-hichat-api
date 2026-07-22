package config

import (
	"fmt"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	ListenOn string

	Mysql struct {
		DataSource string
	} `json:",optional"`

	Cache cache.CacheConf

	// 定时任务配置
	Cron struct {
		// 是否启用秒级精度
		WithSeconds bool
		// 任务并发限制
		MaxConcurrency int
		// 任务超时时间(秒)
		TaskTimeout              int
		InvitationExpirationSpec string `json:",optional"`
		BatchSize                int    `json:",optional"`
		LockTTLSeconds           int    `json:",optional"`
	}

	// 各个微服务的RPC配置
	SocialRpc     zrpc.RpcClientConf
	RpcAuthSecret string `json:",optional"`
}

const defaultTaskTimeoutSeconds = 300
const rpcAuthSecretEnv = "HICHAT_SOCIAL_RPC_AUTH_SECRET"

func (c *Config) ApplyDefaultsAndValidate() error {
	if c.Cron.TaskTimeout <= 0 {
		c.Cron.TaskTimeout = defaultTaskTimeoutSeconds
	}
	if c.Cron.InvitationExpirationSpec == "" {
		return nil
	}
	if _, err := c.LoadRPCAuthSecret(); err != nil {
		return err
	}
	if c.Cron.BatchSize == 0 {
		c.Cron.BatchSize = 200
	} else if c.Cron.BatchSize < 1 {
		c.Cron.BatchSize = 1
	} else if c.Cron.BatchSize > 500 {
		c.Cron.BatchSize = 500
	}
	if c.Cron.LockTTLSeconds <= 0 {
		c.Cron.LockTTLSeconds = c.Cron.TaskTimeout + 60
	}
	if c.Cron.LockTTLSeconds <= c.Cron.TaskTimeout {
		return fmt.Errorf("Cron.LockTTLSeconds must exceed Cron.TaskTimeout")
	}
	if len(c.Cache) == 0 {
		return fmt.Errorf("Cache must contain a Redis node when invitation expiration is enabled")
	}
	var parser cron.Parser
	if c.Cron.WithSeconds {
		parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	} else {
		parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	}
	if _, err := parser.Parse(c.Cron.InvitationExpirationSpec); err != nil {
		return fmt.Errorf("invalid Cron.InvitationExpirationSpec: %w", err)
	}
	return nil
}

func (c Config) LoadRPCAuthSecret() (string, error) {
	secret := c.RpcAuthSecret
	if value, ok := os.LookupEnv(rpcAuthSecretEnv); ok && value != "" {
		secret = value
	}
	if len(secret) < 32 {
		return "", fmt.Errorf("social RPC auth secret must be at least 32 bytes")
	}
	return secret, nil
}

func (c Config) TaskTimeout() time.Duration {
	seconds := c.Cron.TaskTimeout
	if seconds <= 0 {
		seconds = defaultTaskTimeoutSeconds
	}
	return time.Duration(seconds) * time.Second
}
