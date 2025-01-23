package network

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestScheduleInterval(t *testing.T) {
	t.Run("schedule interval job", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		require.NotNil(t, h)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		m := &manager{
			logger: slog.New(h),
			cfg: &Config{Sources: []*Source{
				{
					Name: "test schedule interval",
					Requests: []*Request{
						{Name: "request1", URI: ts.URL, Timeout: time.Second, Interval: 100 * time.Millisecond},
					},
				},
			}},
		}

		m.ctx, m.cancel = context.WithTimeout(context.Background(), 250*time.Millisecond)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			m.schedule()
			wg.Done()
		}(&wg)

		wg.Wait()

		assert.Equal(t, h.Count(), 6)
	})
}

func TestScheduleSchedule(t *testing.T) {
	t.Run("schedule cron job", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		require.NotNil(t, h)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		m := &manager{
			logger: slog.New(h),
			cfg: &Config{Sources: []*Source{
				{
					Name: "test schedule interval",
					Requests: []*Request{
						{Name: "request1", URI: ts.URL, Timeout: time.Second, Schedule: "*/5 8-18 * * 0,6"},
					},
				},
			}},
		}

		m.ctx, m.cancel = context.WithTimeout(context.Background(), 250*time.Millisecond)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			m.schedule()
			wg.Done()
		}(&wg)

		wg.Wait()

		assert.Equal(t, h.Count(), 2)
	})
}

func TestScheduleFail(t *testing.T) {
	t.Run("schedule fail", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		require.NotNil(t, h)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		m := &manager{
			logger: slog.New(h),
			cfg: &Config{Sources: []*Source{
				{
					Name: "test schedule interval",
					Requests: []*Request{
						{Name: "request1", URI: ts.URL, Timeout: time.Second},
					},
				},
			}},
		}

		m.ctx, m.cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			m.schedule()
			wg.Done()
		}(&wg)

		wg.Wait()

		assert.Equal(t, h.Count(), 3)
	})
}
