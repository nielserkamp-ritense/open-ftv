package cerbos

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerbos/cerbos-sdk-go/testutil"
	"github.com/stretchr/testify/require"
)

const (
	adminUser = "cerbos"
	adminPswd = "cerbos"
)

func newService(t *testing.T, cfgPath string) *testutil.CerbosServerInstance {
	launcher, err := testutil.NewCerbosServerLauncher()
	require.NoError(t, err)

	path, err2 := filepath.Abs(cfgPath)
	require.NoError(t, err2)

	s, err3 := launcher.Launch(testutil.LaunchConf{
		ConfFilePath: path,
		Cmd:          []string{"server"},
	})
	require.NoError(t, err3)
	require.NotNil(t, s)

	t.Cleanup(func() { _ = s.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, s.WaitForReady(ctx), "Server failed to start")

	return s
}
