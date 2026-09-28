package slog

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsoleHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		log  func(l *slog.Logger)
		want string
	}{
		{
			name: "message with attributes",
			log: func(l *slog.Logger) {
				l.Info("http service (main) starting", "host", "0.0.0.0", "port", 9020)
			},
			want: "INFO http service (main) starting host=0.0.0.0 port=9020\n",
		},
		{
			name: "message without attributes",
			log:  func(l *slog.Logger) { l.Warn("stopping") },
			want: "WARN stopping\n",
		},
		{
			name: "record below the level is dropped",
			log:  func(l *slog.Logger) { l.Debug("not written", "key", "value") },
			want: "",
		},
		{
			name: "values are quoted when needed",
			log: func(l *slog.Logger) {
				l.Error("failed to start", "error", "connection refused", "detail", "")
			},
			want: `ERROR failed to start error="connection refused" detail=""` + "\n",
		},
		{
			name: "multi-line values stay on one line",
			log: func(l *slog.Logger) {
				l.Error("failed to start", "error", "first\nsecond")
			},
			want: `ERROR failed to start error="first\nsecond"` + "\n",
		},
		{
			name: "handler attributes come first",
			log: func(l *slog.Logger) {
				l.With("app", "manager").Info("started", "port", 9020)
			},
			want: "INFO started app=manager port=9020\n",
		},
		{
			name: "groups prefix the keys",
			log: func(l *slog.Logger) {
				l.WithGroup("db").With("host", "localhost").Info("connected", "port", 5432)
			},
			want: "INFO connected db.host=localhost db.port=5432\n",
		},
		{
			name: "an inlined group in a group gets the prefix once",
			log: func(l *slog.Logger) {
				l.WithGroup("db").With(slog.Group("", "host", "localhost")).Info("connected")
			},
			want: "INFO connected db.host=localhost\n",
		},
		{
			name: "group attributes are flattened",
			log: func(l *slog.Logger) {
				l.Info("connected", slog.Group("db", "host", "localhost", "port", 5432))
			},
			want: "INFO connected db.host=localhost db.port=5432\n",
		},
		{
			name: "empty groups are skipped",
			log:  func(l *slog.Logger) { l.Info("connected", slog.Group("db")) },
			want: "INFO connected\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.Buffer{}
			tc.log(slog.New(NewConsoleHandler(&buf, nil)))

			assert.Equal(t, tc.want, buf.String())
		})
	}
}

func TestConsoleHandlerDerivedHandlersAreIndependent(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	parent := slog.New(NewConsoleHandler(&buf, nil))

	child := parent.With("app", "manager").WithGroup("db")
	child.Info("connected", "port", 5432)
	parent.Info("started", "port", 9020)

	assert.Equal(t, "INFO connected app=manager db.port=5432\nINFO started port=9020\n", buf.String())
}

func TestConsoleHandlerAddsSource(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	slog.New(NewConsoleHandler(&buf, &slog.HandlerOptions{AddSource: true})).Info("started", "port", 9020)

	assert.Regexp(t, `^INFO started port=9020 source=slog/console_test\.go:\d+\n$`, buf.String())
}

func TestConsoleHandlerResolvesValues(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	slog.New(NewConsoleHandler(&buf, nil)).Info("authenticated", "user", testValuer{})

	assert.Equal(t, "INFO authenticated user=redacted\n", buf.String())
}

func TestConsoleHandlerIsConcurrencySafe(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	logger := slog.New(NewConsoleHandler(&buf, nil))

	done := make(chan struct{})

	for i := 0; i < 8; i++ {
		go func() {
			logger.With("worker", "x").Info("working", "key", "value")

			done <- struct{}{}
		}()
	}

	for i := 0; i < 8; i++ {
		<-done
	}

	lines := bytes.Split(bytes.TrimSuffix(buf.Bytes(), []byte("\n")), []byte("\n"))
	require.Len(t, lines, 8)

	for i := range lines {
		assert.Equal(t, "INFO working worker=x key=value", string(lines[i]))
	}
}

type testValuer struct{}

func (testValuer) LogValue() slog.Value { return slog.StringValue("redacted") }
