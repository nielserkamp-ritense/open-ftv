package pip

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestLoadFDS(t *testing.T) {
	_ = os.Chmod("../../../testdata/unittest/fds/test6/ledenlijst.yaml", 0222)
	defer func() {
		_ = os.Chmod("../../../testdata/unittest/fds/test6/ledenlijst.yaml", 0664)
	}()

	testCases := []struct {
		name    string
		path    string
		wantLog int
		want    bool
	}{
		{name: "bad path", path: "/this/is/not/valid/duh", wantLog: 1},
		{name: "good path", path: "../../../testdata/unittest/fds/test1", wantLog: 1, want: true},
		{name: "empty paths", path: "../../../testdata/unittest/fds/test2", wantLog: 1},
		{name: "bad url", path: "../../../testdata/unittest/fds/test3", wantLog: 1},
		{name: "disabled", path: "../../../testdata/unittest/fds/test4", wantLog: 1},
		{name: "not yaml", path: "../../../testdata/unittest/fds/test5", wantLog: 1},
		{name: "read forbidden", path: "../../../testdata/unittest/fds/test6", wantLog: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			p := &pip{ctx: ctx, logger: slog.New(h)}

			p.loadFDS(tc.path)

			time.Sleep(50 * time.Millisecond)
			cancel()

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.want {
				assert.NotNil(t, p.fds)

				f, ok := p.fds.(*fds)
				require.True(t, ok)
				require.NotNil(t, f)
			} else {
				assert.Nil(t, p.fds)
			}
		})
	}
}
