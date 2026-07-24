package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRPCAuthSecret(t *testing.T) {
	t.Setenv(rpcAuthSecretEnv, "")
	_, err := LoadRPCAuthSecret("short")
	require.Error(t, err)
	configured := "0123456789abcdef0123456789abcdef"
	secret, err := LoadRPCAuthSecret(configured)
	require.NoError(t, err)
	require.Equal(t, configured, secret)
	environment := "abcdef0123456789abcdef0123456789"
	t.Setenv(rpcAuthSecretEnv, environment)
	secret, err = LoadRPCAuthSecret(configured)
	require.NoError(t, err)
	require.Equal(t, environment, secret)
}

func TestOptionalRPCAuthSecret(t *testing.T) {
	t.Setenv(rpcAuthSecretEnv, "")
	require.Empty(t, OptionalRPCAuthSecret("short"))
	const secret = "0123456789abcdef0123456789abcdef"
	require.Equal(t, secret, OptionalRPCAuthSecret(secret))
}
