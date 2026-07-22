package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestApplyDefaultsAndValidate(t *testing.T) {
	t.Setenv(rpcAuthSecretEnv, "0123456789abcdef0123456789abcdef")
	tests := []struct {
		name      string
		configure func(*Config)
		wantErr   string
	}{
		{name: "disabled needs no cache"},
		{name: "enabled defaults", configure: func(c *Config) {
			c.Cron.InvitationExpirationSpec = "* * * * *"
			c.Cache = testCacheConf()
		}},
		{name: "ttl must exceed timeout", configure: func(c *Config) {
			c.Cron.InvitationExpirationSpec = "* * * * *"
			c.Cron.TaskTimeout = 60
			c.Cron.LockTTLSeconds = 60
			c.Cache = testCacheConf()
		}, wantErr: "must exceed"},
		{name: "enabled requires cache", configure: func(c *Config) {
			c.Cron.InvitationExpirationSpec = "* * * * *"
		}, wantErr: "Cache must contain"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Config
			if tt.configure != nil {
				tt.configure(&c)
			}
			err := c.ApplyDefaultsAndValidate()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 300*time.Second, c.TaskTimeout())
			if c.Cron.InvitationExpirationSpec != "" {
				require.Equal(t, 200, c.Cron.BatchSize)
				require.Greater(t, c.Cron.LockTTLSeconds, c.Cron.TaskTimeout)
			}
		})
	}
}

func TestApplyDefaultsAndValidateClampsBatch(t *testing.T) {
	t.Setenv(rpcAuthSecretEnv, "0123456789abcdef0123456789abcdef")
	for _, tt := range []struct {
		name string
		in   int
		want int
	}{
		{name: "minimum", in: -1, want: 1},
		{name: "maximum", in: 900, want: 500},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var c Config
			c.Cron.InvitationExpirationSpec = "* * * * *"
			c.Cron.BatchSize = tt.in
			c.Cache = testCacheConf()
			require.NoError(t, c.ApplyDefaultsAndValidate())
			require.Equal(t, tt.want, c.Cron.BatchSize)
		})
	}
}

func TestApplyDefaultsAndValidateCronExpressionAndSecret(t *testing.T) {
	tests := []struct {
		name        string
		withSeconds bool
		spec        string
		secret      string
		wantErr     string
	}{
		{name: "five fields", spec: "*/5 * * * *", secret: "0123456789abcdef0123456789abcdef"},
		{name: "six fields", withSeconds: true, spec: "*/5 * * * * *", secret: "0123456789abcdef0123456789abcdef"},
		{name: "wrong field count", spec: "*/5 * * * * *", secret: "0123456789abcdef0123456789abcdef", wantErr: "invalid Cron"},
		{name: "short secret", spec: "*/5 * * * *", secret: "short", wantErr: "at least 32"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(rpcAuthSecretEnv, "")
			var c Config
			c.Cron.InvitationExpirationSpec = tt.spec
			c.Cron.WithSeconds = tt.withSeconds
			c.RpcAuthSecret = tt.secret
			c.Cache = testCacheConf()
			err := c.ApplyDefaultsAndValidate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestLoadRPCAuthSecretPrefersEnvironment(t *testing.T) {
	environment := "abcdef0123456789abcdef0123456789"
	t.Setenv(rpcAuthSecretEnv, environment)
	secret, err := (Config{RpcAuthSecret: "0123456789abcdef0123456789abcdef"}).LoadRPCAuthSecret()
	require.NoError(t, err)
	require.Equal(t, environment, secret)
}

func testCacheConf() cache.CacheConf {
	return cache.CacheConf{{RedisConf: redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}}}
}
