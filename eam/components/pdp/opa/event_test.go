package opa

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/hooks"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestController_Handle(t *testing.T) {
	p1 := "package authz\ndefault allow = false"

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
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "add - new key",
			policies:   map[string]string{"rego/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "add - duplicate",
			policies:   map[string]string{"rego/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyAdded,
			key2:       "rego/p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "add & replace",
			policies:   map[string]string{"rego/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "rego/p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "replace - not found",
			event1:     models.PolicyReplaced,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "replace - new key",
			policies:   map[string]string{"rego/p1": p1},
			event1:     models.PolicyReplaced,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "replace - duplicate",
			policies:   map[string]string{"rego/p1": p1},
			event1:     models.PolicyReplaced,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "rego/p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "remove - not found",
			event1:     models.PolicyRemoved,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "failed to remove policy",
		},
		{
			name:       "remove - found",
			policies:   map[string]string{"rego/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "rego/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyRemoved,
			key2:       "rego/p1",
			wantLog2:   1,
			logPrefix2: "policy removed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			mem := inmem.New()

			engine, err := sdk.New(context.Background(), sdk.Options{
				RegoVersion:   1,
				ID:            "opa-controller",
				Config:        bytes.NewReader([]byte(cfg)),
				ConsoleLogger: &wrappedLogger{logger: logger},
				Hooks:         hooks.Hooks{},
				Store:         mem,
			})
			require.NoError(t, err)

			c := &controller{
				Base: pdp.NewBase(pdp.WithNameVersion("x", "v1"), pdp.WithLogger(logger)),
				pdp:  engine,
				mem:  mem,
			}

			p := pap.New(nil, logger)

			for key := range tc.policies {
				data := []byte(tc.policies[key])
				parts := strings.Split(key, "/")

				pol, err2 := pap.NewPolicy(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewReader(data))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol)
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
