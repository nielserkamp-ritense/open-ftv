//go:build integration

package cerbos_api

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// TestController_Handle requires a live Cerbos server (it asserts c2.info, which needs a real
// server-info RPC to succeed) so it only runs with `go test -tags=integration`.
func TestController_Handle(t *testing.T) {
	p1 := `{
 "apiVersion": "api.cerbos.dev/v1",
 "rolePolicy": {
  "role": "admin",
  "scopePermissions": "SCOPE_PERMISSIONS_REQUIRE_PARENTAL_CONSENT_FOR_ALLOWS",
  "rules": [
   { "resource": "service:https://inway-fsc-nlx-inway:443/brp-personen", "allowActions": ["POST"]}
  ]
 }
}`

	addr := getAddress()

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
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "add - new key",
			policies:   map[string]string{"cerbos/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "add - duplicate",
			policies:   map[string]string{"cerbos/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyAdded,
			key2:       "cerbos/p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "add & replace",
			policies:   map[string]string{"cerbos/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "cerbos/p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "replace - not found",
			event1:     models.PolicyReplaced,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "failed to get policy",
		},
		{
			name:       "replace - new key",
			policies:   map[string]string{"cerbos/p1": p1},
			event1:     models.PolicyReplaced,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
		},
		{
			name:       "replace - duplicate",
			policies:   map[string]string{"cerbos/p1": p1},
			event1:     models.PolicyReplaced,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "cerbos/p1",
			wantLog2:   1,
			logPrefix2: "policy added/replaced",
		},
		{
			name:       "remove - found",
			policies:   map[string]string{"cerbos/p1": p1},
			event1:     models.PolicyAdded,
			key1:       "cerbos/p1",
			wantLog1:   1,
			logPrefix1: "policy added/replaced",
			event2:     models.PolicyRemoved,
			key2:       "cerbos/p1",
			wantLog2:   1,
			logPrefix2: "policy removed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p, err := pap.New(nil, logger, pap.WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			for key := range tc.policies {
				data := []byte(tc.policies[key])
				parts := strings.Split(key, "/")

				pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewReader(data))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol, identity.NewPrincipal(identity.KindUser, "test"))
				require.NoError(t, err2)
			}

			c := NewController(
				Config{Addr1: addr, Addr2: addr, User: adminUser, Pswd: adminPswd},
				pdp.WithLogger(logger),
				pdp.WithPAP(p),
			)

			c2, ok := c.(*controller)
			require.True(t, ok)
			require.NotNil(t, c2)
			require.NotNilf(t, c2.engine, h.Log())
			require.NotNilf(t, c2.admin, h.Log())
			require.NotNilf(t, c2.info, h.Log())

			h.Clear()
			c2.Handle(tc.event1, tc.key1)

			assert.Equal(t, tc.wantLog1, h.Count())

			if tc.wantLog1 > 0 && tc.logPrefix1 != "" {
				var i int
				h.Iterate(func(_ time.Time, msg string, _ slog.Level) {
					if i == 0 {
						assert.Equalf(t, tc.logPrefix1, msg, h.Log())
					}
					i++
				})
			}

			if tc.event2 > 0 {
				h.Clear()
				c2.Handle(tc.event2, tc.key2)

				assert.Equal(t, tc.wantLog2, h.Count())

				if tc.wantLog2 > 0 && tc.logPrefix2 != "" {
					var i int
					h.Iterate(func(_ time.Time, msg string, _ slog.Level) {
						if i == 0 {
							assert.Equalf(t, tc.logPrefix2, msg, h.Log())
						}
						i++
					})
				}
			}
		})
	}
}
