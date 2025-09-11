package pap

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPAP_NewDeployment(t *testing.T) {
	t.Parallel()

	t.Run("new deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("v1", "merry easter", m, "*SYSTEM*")
		require.NoError(t, err)
		require.NotNil(t, d)

		assert.Equal(t, uint64(1), d.Version())
		assert.Equal(t, bundles.Creating, d.Status())
	})
}

func TestPAP_LastDeployment(t *testing.T) {
	t.Parallel()

	t.Run("last deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("v2", "merry easter", m, "*SYSTEM*")
		require.NoError(t, err)
		require.NotNil(t, d)

		d2, err2 := p.LastDeployment()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.GreaterOrEqual(t, d2.Status(), bundles.Creating)
	})
}

func TestPAP_RestartDeployment(t *testing.T) {
	t.Parallel()

	t.Run("restart deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("v3", "happy christmas", m, "*SYSTEM*")
		require.NoError(t, err)
		require.NotNil(t, d)

		cancel()

		p.RestartDeployment(m)

		time.Sleep(25 * time.Millisecond)

		d2, err2 := p.LastDeployment()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.GreaterOrEqual(t, d2.Status(), bundles.Creating)
	})
}

func TestPAP_ReadDeployment(t *testing.T) {
	t.Parallel()

	t.Run("read deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		_, err := p.NewDeployment("v4", "merry easter", m, "*SYSTEM*")
		require.NoError(t, err)

		d2, err2 := p.ReadDeployment(1)
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.GreaterOrEqual(t, d2.Status(), bundles.Creating)

		d3, err3 := p.ReadDeployment(2)
		require.Error(t, err3)
		require.Nil(t, d3)
	})
}

func TestPAP_ListDeployments(t *testing.T) {
	t.Parallel()

	t.Run("list deployments", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("v5", "merry easter", m, "*SYSTEM*")
		require.NoError(t, err)
		require.NotNil(t, d)

		list, err2 := p.ListDeployments()
		require.NoError(t, err2)
		require.NotNil(t, list)
		assert.Equal(t, 1, len(list))
	})
}
