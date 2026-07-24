package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	dir := t.TempDir()
	content := []byte("mysql:\n  host: 127.0.0.1\n  port: 3307\n  name: hichat2\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config-local.yaml"), content, 0o600))

	InitConfig("local", dir)
	require.Equal(t, "127.0.0.1", ServiceConf.DB.Host)
	require.Equal(t, 3307, ServiceConf.DB.Port)
	require.Equal(t, "hichat2", ServiceConf.DB.Name)
}
