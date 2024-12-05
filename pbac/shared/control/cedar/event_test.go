package cedar

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestController_Handle(t *testing.T) {
	p1 := `permit (
    principal is User,
    action == Action::"POST",
    resource is Service
);`

	testCases := []struct {
		name       string
		policies   map[string]string
		event1     models.EventType
		key1       string
		wantLog1   int
		logPrefix1 string
		event2     models.EventType
		key2       string
		wantLog2   int
		logPrefix2 string
	}{
		{
			name:       "add - not found",
			event1:     models.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "add - new key",
			policies:   map[string]string{"p1": p1},
			event1:     models.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "add - duplicate",
			policies:   map[string]string{"p1": p1},
			event1:     models.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyAdded,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "add & replace",
			policies:   map[string]string{"p1": p1},
			event1:     models.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "replace - not found",
			event1:     models.PolicyReplaced,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "replace - new key",
			policies:   map[string]string{"p1": p1},
			event1:     models.PolicyReplaced,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "replace - duplicate",
			policies:   map[string]string{"p1": p1},
			event1:     models.PolicyReplaced,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "remove - found",
			policies:   map[string]string{"p1": p1},
			event1:     models.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyRemoved,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy removed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			pdp := cedar.NewPolicySet()
			require.NotNil(t, pdp)

			c := &controller{
				Base: control.NewBase("x", "v1", logger, nil),
				pdp:  pdp,
			}

			p := pap.New(nil, logger, nil)
			for key := range tc.policies {
				pol := []byte(tc.policies[key])
				err2 := p.Add(key, bytes.NewReader(pol))
				require.NoError(t, err2)
			}

			c.SetPAP(p)

			h.Clear()
			c.Handle(tc.event1, tc.key1)

			assert.Equal(t, tc.wantLog1, h.Count())

			if tc.wantLog1 > 0 && tc.logPrefix1 != "" {
				var i int
				h.Iterate(func(_ time.Time, msg string, _ slog.Level) {
					if i == 0 {
						assert.Equal(t, tc.logPrefix1, msg)
					}
					i++
				})
			}

			if tc.event2 > 0 {
				h.Clear()
				c.Handle(tc.event2, tc.key2)

				assert.Equal(t, tc.wantLog2, h.Count())

				if tc.wantLog2 > 0 && tc.logPrefix2 != "" {
					var i int
					h.Iterate(func(_ time.Time, msg string, _ slog.Level) {
						if i == 0 {
							assert.Equal(t, tc.logPrefix2, msg)
						}
						i++
					})
				}
			}
		})
	}
}
