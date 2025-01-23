package network

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewManager(t *testing.T) {
	d := t.TempDir()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data1 := fmt.Sprintf(`sources:
  - name: source1
    requests:
      - name: request1
        uri: %s
        timeout: 1s
        interval: 100ms
`, ts.URL)

	testCases := []struct {
		name      string
		data      string
		timeout   time.Duration
		wantErr   bool
		wantCount int
	}{
		{
			name:      "bad config",
			data:      "oops",
			timeout:   10 * time.Millisecond,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "interval job",
			data:      data1,
			timeout:   250 * time.Millisecond,
			wantCount: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			require.NotNil(t, h)

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			path := filepath.Join(d, "test.yaml")

			f, err := os.Create(path)
			require.NoError(t, err)

			_, err = f.Write([]byte(tc.data))
			require.NoError(t, err)

			err = f.Close()
			require.NoError(t, err)

			m, err2 := NewManager(ctx, path, slog.New(h), &nilGetter{})
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, m)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, m)
			}

			select {
			case <-ctx.Done():
				// make sure we catch the last log line!
				time.Sleep(10 * time.Millisecond)
			}

			assert.Equal(t, tc.wantCount, h.Count())
		})
	}
}
