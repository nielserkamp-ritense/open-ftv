package pip

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/fds/ledenlijst"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewFDS(t *testing.T) {
	t.Run("new FDS", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := json.Marshal(leden)
			require.NoError(t, err)
			_, _ = w.Write(b)
		}))
		defer svr.Close()

		u, err := url.Parse(svr.URL)
		require.NoError(t, err)

		h := slog2.NewDummyHandler(slog.LevelDebug)

		f := newFDS(ctx, slog.New(h), u, time.Minute)
		require.NotNil(t, f)

		time.Sleep(50 * time.Millisecond)
		cancel()

		assert.Zero(t, h.Count())

		f2, ok := f.(*fds)
		require.True(t, ok)
		require.NotNil(t, f2)

		assert.Equal(t, 1, len(f2.leden))

		m := f2.leden["1"]
		assert.Equal(t, MaturityLevel(3), m)
	})
}

var leden = []ledenlijst.Organization{
	{Attributes: map[string]any{"isMember": true, "maturity": 3}, Id: "l1", Oin: "1"},
}

func TestNewFDS_BadStatus(t *testing.T) {
	t.Run("new FDS - bad status", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer svr.Close()

		u, err := url.Parse(svr.URL)
		require.NoError(t, err)

		h := slog2.NewDummyHandler(slog.LevelDebug)

		f := newFDS(ctx, slog.New(h), u, time.Minute)
		require.NotNil(t, f)

		time.Sleep(50 * time.Millisecond)
		cancel()

		assert.Equal(t, 1, h.Count())

		f2, ok := f.(*fds)
		require.True(t, ok)
		require.NotNil(t, f2)

		assert.Empty(t, f2.leden)
	})
}

func TestNewFDS_InvalidJSON(t *testing.T) {
	t.Run("new FDS - bad response", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not a json string"))
		}))
		defer svr.Close()

		u, err := url.Parse(svr.URL)
		require.NoError(t, err)

		h := slog2.NewDummyHandler(slog.LevelDebug)

		f := newFDS(ctx, slog.New(h), u, time.Minute)
		require.NotNil(t, f)

		time.Sleep(50 * time.Millisecond)
		cancel()

		assert.Equal(t, 1, h.Count())

		f2, ok := f.(*fds)
		require.True(t, ok)
		require.NotNil(t, f2)

		assert.Empty(t, f2.leden)
	})
}

func TestNewFDS_BadURL(t *testing.T) {
	t.Run("new FDS - bad url", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		u := &url.URL{
			Scheme: "x",
			Host:   string([]byte{0, 1, 2}),
			Path:   "\\\\",
		}

		h := slog2.NewDummyHandler(slog.LevelDebug)

		f := newFDS(ctx, slog.New(h), u, time.Minute)
		require.NotNil(t, f)

		time.Sleep(50 * time.Millisecond)
		cancel()

		assert.Equal(t, 1, h.Count())

		f2, ok := f.(*fds)
		require.True(t, ok)
		require.NotNil(t, f2)

		assert.Empty(t, f2.leden)
	})
}

func TestNewFDS_RequestFail(t *testing.T) {
	t.Run("new FDS - request fail", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("oops")
		}))
		defer svr.Close()

		u, err := url.Parse(svr.URL)
		require.NoError(t, err)

		h := slog2.NewDummyHandler(slog.LevelDebug)

		f := newFDS(ctx, slog.New(h), u, time.Minute)
		require.NotNil(t, f)

		time.Sleep(50 * time.Millisecond)
		cancel()

		assert.Equal(t, 1, h.Count())

		f2, ok := f.(*fds)
		require.True(t, ok)
		require.NotNil(t, f2)

		assert.Empty(t, f2.leden)
	})
}
