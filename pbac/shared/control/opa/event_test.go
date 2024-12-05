package opa

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/hooks"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestController_Handle(t *testing.T) {
	p1 := "package authz\ndefault allow = false"

	testCases := []struct {
		name       string
		policies   map[string]string
		event1     standards.EventType
		key1       string
		wantLog1   int
		logPrefix1 string
		event2     standards.EventType
		key2       string
		wantLog2   int
		logPrefix2 string
	}{
		{
			name:       "add - not found",
			event1:     standards.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "add - new key",
			policies:   map[string]string{"p1": p1},
			event1:     standards.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "add - duplicate",
			policies:   map[string]string{"p1": p1},
			event1:     standards.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     standards.PolicyAdded,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "add & replace",
			policies:   map[string]string{"p1": p1},
			event1:     standards.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     standards.PolicyReplaced,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "replace - not found",
			event1:     standards.PolicyReplaced,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "replace - new key",
			policies:   map[string]string{"p1": p1},
			event1:     standards.PolicyReplaced,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "replace - duplicate",
			policies:   map[string]string{"p1": p1},
			event1:     standards.PolicyReplaced,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     standards.PolicyReplaced,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "remove - not found",
			event1:     standards.PolicyRemoved,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "failed to remove policy",
		},
		{
			name:       "remove - found",
			policies:   map[string]string{"p1": p1},
			event1:     standards.PolicyAdded,
			key1:       "p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     standards.PolicyRemoved,
			key2:       "p1",
			wantLog2:   1,
			logPrefix2: "policy removed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			mem := inmem.New()

			pdp, err := sdk.New(context.Background(), sdk.Options{
				RegoVersion:   1,
				ID:            "opa-controller",
				Config:        bytes.NewReader([]byte(cfg)),
				ConsoleLogger: &wrappedLogger{logger: logger},
				Hooks:         hooks.Hooks{},
				Store:         mem,
			})
			require.NoError(t, err)

			c := &controller{
				Base: control.NewBase("x", "v1", logger, nil),
				pdp:  pdp,
				mem:  mem,
				ctx:  context.Background(),
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
