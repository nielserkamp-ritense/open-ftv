package migrate

import (
	"log/slog"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/assert"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestWrapper(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		verbose bool
		f       func(w migrate.Logger)
		want    int
	}{
		{
			name: "none",
			f:    func(w migrate.Logger) {},
		},
		{
			name:    "one debug; not verbose",
			verbose: false,
			f: func(w migrate.Logger) {
				if w.Verbose() {
					w.Printf("hello %s %d", "world", 123)
				}
			},
			want: 0,
		},
		{
			name:    "few debug; verbose",
			verbose: true,
			f: func(w migrate.Logger) {
				if w.Verbose() {
					w.Printf("hello %s %d", "world", 123)
					w.Printf("hello %s %d", "mars", 1)
					w.Printf("hello %s %d", "jupter", 2)
				}
			},
			want: 3,
		},
		{
			name:    "few; not verbose",
			verbose: false,
			f: func(w migrate.Logger) {
				w.Printf("hello %s %d", "mars", 1)
				w.Printf("hello %s %d", "world", 2)
				w.Printf("hello %s %d", "venus", 3)
				w.Printf("hello %s %d", "pluto", 4)
			},
			want: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			w := &wrapper{logger: logger, verbose: tc.verbose}
			tc.f(w)

			assert.Equal(t, tc.want, h.Count())
		})
	}
}
