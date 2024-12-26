package openfga

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestController_HandleModel(t *testing.T) {
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
			name:       "add model - not found",
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "failed to get file",
		},
		{
			name:       "replace model - not found",
			event1:     models.PolicyReplaced,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "failed to get file",
		},
		{
			name:       "add model - bad store name",
			policies:   map[string]string{"s3.model": store1},
			event1:     models.PolicyAdded,
			key1:       "s3.model",
			wantLog1:   1,
			logPrefix1: "failed to create store",
		},
		{
			name:       "add model - bad model",
			policies:   map[string]string{"store2.model": store2},
			event1:     models.PolicyAdded,
			key1:       "store2.model",
			wantLog1:   1,
			logPrefix1: "failed to compile model",
		},
		{
			name:       "add model - new key",
			policies:   map[string]string{"store1.model": store1},
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
		},
		{
			name:       "add model - duplicate",
			policies:   map[string]string{"store1.model": store1},
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
			event2:     models.PolicyAdded,
			key2:       "store1.model",
			wantLog2:   1,
			logPrefix2: "model added/replaced",
		},
		{
			name:       "add & replace model",
			policies:   map[string]string{"store1.model": store1},
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "store1.model",
			wantLog2:   1,
			logPrefix2: "model added/replaced",
		},
		{
			name:       "replace model - new key",
			policies:   map[string]string{"store1.model": store1},
			event1:     models.PolicyReplaced,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
		},
		{
			name:       "replace model - duplicate",
			policies:   map[string]string{"store1.model": store1},
			event1:     models.PolicyReplaced,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "store1.model",
			wantLog2:   1,
			logPrefix2: "model added/replaced",
		},
		{
			name:       "remove model - found",
			policies:   map[string]string{"store1.model": store1},
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
			event2:     models.PolicyRemoved,
			key2:       "store1.model",
			wantLog2:   1,
			logPrefix2: "store removed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			engine, err := server.NewServerWithOpts(
				server.WithDatastore(memory.New()),
				server.WithLogger(newZapper(logger)),
			)
			require.NoError(t, err)
			require.NotNil(t, engine)

			c := &controller{
				Base:   pdp.NewBase(pdp.WithNameVersion("x", "v1"), pdp.WithLogger(logger)),
				engine: engine,
				stores: make(map[string]*details),
			}

			p := pap.New(nil, logger, nil)
			for id := range tc.policies {
				data := []byte(tc.policies[id])

				pol, err2 := pap.NewPolicy(&policies.Policy{Id: id}, bytes.NewReader(data))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Add(pol)
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

func TestController_HandleRelations(t *testing.T) {
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
		event3     models.EventType
		key3       string
		wantLog3   int
		logPrefix3 string
	}{
		{
			name:       "add relations - not found",
			event2:     models.PolicyAdded,
			key2:       "store1.relations",
			wantLog2:   1,
			logPrefix2: "failed to get file",
		},
		{
			name:       "replace relations - not found",
			event2:     models.PolicyReplaced,
			key2:       "store1.relations",
			wantLog2:   1,
			logPrefix2: "failed to get file",
		},
		{
			name:       "add relation - bad store name",
			policies:   map[string]string{"s3.relations": relations1},
			event2:     models.PolicyAdded,
			key2:       "s3.relations",
			wantLog2:   1,
			logPrefix2: "failed to create store",
		},
		{
			name:       "add relation - bad",
			policies:   map[string]string{"store2.relations": relations2},
			event2:     models.PolicyAdded,
			key2:       "store2.relations",
			wantLog2:   1,
			logPrefix2: "failed to decode relations",
		},
		{
			name:       "add relation - new key",
			policies:   map[string]string{"store1.model": store1, "store1.relations": relations1},
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
			event2:     models.PolicyAdded,
			key2:       "store1.relations",
			wantLog2:   1,
			logPrefix2: "relations added/replaced",
		},
		// {
		// 	name:       "add relation - duplicate",
		// 	policies:   map[string]string{"store1.model": store1, "store1.relations": relations1},
		// 	event1:     models.PolicyAdded,
		// 	key1:       "store1.model",
		// 	wantLog1:   1,
		// 	logPrefix1: "model added/replaced",
		// 	event2:     models.PolicyAdded,
		// 	key2:       "store1.relations",
		// 	wantLog2:   1,
		// 	logPrefix2: "relations added/replaced",
		// 	event3:     models.PolicyAdded,
		// 	key3:       "store1.relations",
		// 	wantLog3:   1,
		// 	logPrefix3: "relations added/replaced",
		// },
		// {
		// 	name:       "add & replace relation",
		// 	policies:   map[string]string{"store1.model": store1, "store1.relations": relations1},
		// 	event1:     models.PolicyAdded,
		// 	key1:       "store1.model",
		// 	wantLog1:   1,
		// 	logPrefix1: "model added/replaced",
		// 	event2:     models.PolicyAdded,
		// 	key2:       "store1.relations",
		// 	wantLog2:   1,
		// 	logPrefix2: "relations added/replaced",
		// 	event3:     models.PolicyReplaced,
		// 	key3:       "store1.relations",
		// 	wantLog3:   1,
		// 	logPrefix3: "relations added/replaced",
		// },
		{
			name:       "replace relation - new key",
			policies:   map[string]string{"store1.model": store1, "store1.relations": relations1},
			event1:     models.PolicyAdded,
			key1:       "store1.model",
			wantLog1:   1,
			logPrefix1: "model added/replaced",
			event2:     models.PolicyReplaced,
			key2:       "store1.relations",
			wantLog2:   1,
			logPrefix2: "relations added/replaced",
		},
		// {
		// 	name:       "replace relation - duplicate",
		// 	policies:   map[string]string{"store1.model": store1, "store1.relations": relations1},
		// 	event1:     models.PolicyAdded,
		// 	key1:       "store1.model",
		// 	wantLog1:   1,
		// 	logPrefix1: "model added/replaced",
		// 	event2:     models.PolicyReplaced,
		// 	key2:       "store1.relations",
		// 	wantLog2:   1,
		// 	logPrefix2: "relations added/replaced",
		// 	event3:     models.PolicyReplaced,
		// 	key3:       "store1.relations",
		// 	wantLog3:   1,
		// 	logPrefix3: "relations added/replaced",
		// },
		// {
		// 	name:       "remove relation - found",
		// 	policies:   map[string]string{"store1.model": store1, "store1.relations": relations1},
		// 	event1:     models.PolicyAdded,
		// 	key1:       "store1.model",
		// 	wantLog1:   1,
		// 	logPrefix1: "model added/replaced",
		// 	event2:     models.PolicyAdded,
		// 	key2:       "store1.relations",
		// 	wantLog2:   1,
		// 	logPrefix2: "relations added/replaced",
		// 	event3:     models.PolicyRemoved,
		// 	key3:       "store1.relations",
		// 	wantLog3:   1,
		// 	logPrefix3: "relations removed",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			engine, err := server.NewServerWithOpts(
				server.WithDatastore(memory.New()),
				server.WithLogger(newZapper(logger)),
			)
			require.NoError(t, err)
			require.NotNil(t, engine)

			c := &controller{
				Base:   pdp.NewBase(pdp.WithNameVersion("x", "v1"), pdp.WithLogger(logger)),
				engine: engine,
				stores: make(map[string]*details),
			}

			p := pap.New(nil, logger, nil)
			for id := range tc.policies {
				data := []byte(tc.policies[id])

				pol, err2 := pap.NewPolicy(&policies.Policy{Id: id}, bytes.NewReader(data))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Add(pol)
				require.NoError(t, err2)
			}

			c.SetPAP(p)

			if tc.event1 > 0 {
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

			if tc.event3 > 0 {
				h.Clear()
				c.Handle(tc.event3, tc.key3)

				assert.Equal(t, tc.wantLog3, h.Count())

				if tc.wantLog3 > 0 && tc.logPrefix3 != "" {
					var i int
					h.Iterate(func(_ time.Time, msg string, _ slog.Level) {
						if i == 0 {
							assert.Equal(t, tc.logPrefix3, msg)
						}
						i++
					})
				}
			}
		})
	}
}

const (
	store1 = `
model
  schema 1.1

type doelbinding

type service
  relations
    define call: [doelbinding]
`
	store2 = `model bad`

	relations1 = `[
 {"subject":{"type":"doelbinding","id":"burgerzaken"},"predicate":"call","object":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}},
 {"subject":{"type":"doelbinding","id":"subsidies"},"predicate":"call","object":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}
]`
	relations2 = `not json`
)
