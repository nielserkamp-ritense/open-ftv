package openfga_embedded

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewZapper(t *testing.T) {
	t.Parallel()

	t.Run("test zapper", func(t *testing.T) {
		f1 := zap.Float64("f", 123.456)
		s1 := zap.String("k", "v")

		h := slog2.NewDummyHandler(slog.LevelDebug)
		z := newZapper(slog.New(h))
		require.NotNil(t, z)

		z.Debug("debug1")
		z.Info("info1", s1, f1)
		z.Warn("warn1")
		z.Error("error1", f1)
		z.Panic("panic1")
		z.Fatal("fatal1", s1)
		assert.Equal(t, 6, h.Count())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		z.DebugWithContext(ctx, "debug2", f1, s1, f1)
		z.InfoWithContext(ctx, "info2")
		z.WarnWithContext(ctx, "warn2", s1)
		z.ErrorWithContext(ctx, "error2")
		z.PanicWithContext(ctx, "panic2", f1)
		z.FatalWithContext(ctx, "fatal2")
		assert.Equal(t, 12, h.Count())

		z2 := z.With(f1, s1)

		z2.Debug("debug3")
		z2.Info("info3", f1)
		z2.Warn("warn3")
		z2.Error("error3")
		z2.Panic("panic3", s1)
		z2.Fatal("fatal3")
		assert.Equal(t, 18, h.Count())
	})
}
