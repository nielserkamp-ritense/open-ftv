package pip

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

func TestProcessHeaders(t *testing.T) {
	testCases := []struct {
		name        string
		headers     map[string][]string
		wantURI     string
		wantCT      string
		wantAuth    bool
		wantFSC     bool
		wantKey     string
		wantAID     string
		wantCU      string
		wantGS      string
		wantDB      string
		wantZT      string
		wantTaak    string
		wantFwd     string
		wantHeaders map[string]string
	}{
		{
			name: "empty",
		},
		{
			name:    "content-type",
			headers: map[string][]string{"content-type": {"application/json"}},
			wantCT:  "application/json",
		},
		{
			name:    "Api-Key",
			headers: map[string][]string{"Api-Key": {"123456"}},
			wantKey: "123456",
		},
		{
			name:    "APIkey",
			headers: map[string][]string{"APIkey": {"123456"}},
			wantKey: "123456",
		},
		{
			name:    "x-apikey",
			headers: map[string][]string{"x-apikey": {"123456"}},
			wantKey: "123456",
		},
		{
			name:    "X-API-KEY",
			headers: map[string][]string{"X-API-KEY": {"123456"}},
			wantKey: "123456",
		},
		{
			name:    "Activity-ID",
			headers: map[string][]string{"X-Dpl-Rva-Activity-Id": {"54f6a690-d7f5-4dc9-9ec9-af00e336bf6a"}},
			wantAID: "54f6a690-d7f5-4dc9-9ec9-af00e336bf6a",
		},
		{
			name:    "Core-User",
			headers: map[string][]string{"X-Dpl-Core-User": {"Alice"}},
			wantCU:  "Alice",
		},
		{
			name:    "GrondSlag",
			headers: map[string][]string{"GrondSlag": {"WetBRP"}},
			wantGS:  "WetBRP",
		},
		{
			name:    "doelBinding",
			headers: map[string][]string{"doelBinding": {"mijnzaak"}},
			wantDB:  "mijnzaak",
		},
		{
			name:    "ZaakType",
			headers: map[string][]string{"ZaakType": {"subsidie"}},
			wantZT:  "subsidie",
		},
		{
			name:     "TAAK",
			headers:  map[string][]string{"TAAK": {"controle"}},
			wantTaak: "controle",
		},
		{
			name:    "X-Forwarded-For",
			headers: map[string][]string{"X-Forwarded-For": {"192.168.0.1"}},
			wantFwd: "192.168.0.1",
		},
		{
			name:    "Forwarded",
			headers: map[string][]string{"X-Forwarded-For": {"192.168.0.19"}},
			wantFwd: "192.168.0.19",
		},
		{
			name:    "X-Forwarded-For & Forwarded",
			headers: map[string][]string{"Forwarded": {"127.0.0.1"}, "X-Forwarded-For": {"192.168.0.11"}},
			wantFwd: "192.168.0.11",
		},
		{
			name:        "one other",
			headers:     map[string][]string{"Vessel": {"cup-of-tea"}},
			wantHeaders: map[string]string{"Vessel": "cup-of-tea"},
		},
		{
			name:        "few other",
			headers:     map[string][]string{"Vessel": {"cup-of-tea"}, "parents": {"mom", "dad"}},
			wantHeaders: map[string]string{"Vessel": "cup-of-tea", "parents": "mom,dad"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &pip{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

			a := models.NewAttributeSet()
			require.NotNil(t, a)

			req := &components.Request{Headers: tc.headers}
			newURI := p.testHeaders(req, a)

			if tc.wantURI != "" {
				assert.Equal(t, tc.wantURI, newURI)
			} else {
				assert.Empty(t, newURI)
			}

			if tc.wantCT != "" {
				assert.Equal(t, tc.wantCT, a.GetAttributeValue(models.AttrContentType))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrContentType))
			}

			if tc.wantAuth {
				assert.NotNil(t, a.GetAttributeValue(models.AttrJWT))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrJWT))
			}

			if tc.wantFSC {
				assert.NotNil(t, a.GetAttributeValue(models.AttrFSC))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrFSC))
			}

			if tc.wantKey != "" {
				assert.Equal(t, tc.wantKey, a.GetAttributeValue(models.AttrAPIKey))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrAPIKey))
			}

			if tc.wantAID != "" {
				assert.Equal(t, tc.wantAID, a.GetAttributeValue(models.AttrRvvaID))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrRvvaID))
			}

			if tc.wantCU != "" {
				assert.Equal(t, tc.wantCU, a.GetAttributeValue(models.AttrCoreUser))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrCoreUser))
			}

			if tc.wantGS != "" {
				assert.Equal(t, tc.wantGS, a.GetAttributeValue(models.AttrGrondslag))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrGrondslag))
			}

			if tc.wantDB != "" {
				assert.Equal(t, tc.wantDB, a.GetAttributeValue(models.AttrDoelbinding))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrDoelbinding))
			}

			if tc.wantZT != "" {
				assert.Equal(t, tc.wantZT, a.GetAttributeValue(models.AttrZaakType))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrZaakType))
			}

			if tc.wantTaak != "" {
				assert.Equal(t, tc.wantTaak, a.GetAttributeValue(models.AttrTaak))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrTaak))
			}

			if tc.wantFwd != "" {
				assert.Equal(t, tc.wantFwd, a.GetAttributeValue(models.AttrClientIP))
			} else {
				assert.Nil(t, a.GetAttributeValue(models.AttrClientIP))
			}

			other, ok := a.GetAttributeValue(models.AttrHeaders).(map[string]string)
			require.True(t, ok)
			require.NotNil(t, other)

			if tc.wantHeaders != nil {
				assert.EqualValues(t, tc.wantHeaders, other)
			} else {
				assert.Empty(t, other)
			}
		})
	}
}

func TestProcessActivityID(t *testing.T) {
	testCases := []struct {
		name string
		id   string
		want any
	}{
		{
			name: "empty",
		},
		{
			name: "not found",
			id:   "abc",
		},
		{
			name: "found, invalid",
			id:   "bad",
		},
		{
			name: "found, valid",
			id:   "good",
			want: "zorgtoeslag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))

			p := &pip{
				logger:   logger,
				entities: New(Config{Store: "../../../testdata/unittest/pip2", Recurse: true, Logger: logger}),
			}

			a := models.NewAttributeSet()
			require.NotNil(t, a)

			p.convertActivityID(tc.id, a)
			assert.Equal(t, tc.want, a.GetAttributeValue(models.AttrDoelbinding))
		})
	}
}
