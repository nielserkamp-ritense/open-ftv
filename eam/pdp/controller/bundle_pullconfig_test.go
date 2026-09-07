package controller

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

const pullPollTimeout = 2 * time.Second

const pullPollInterval = 10 * time.Millisecond

// TestBase_NewBundle_keepsPullConfigData is an integration test covering the interaction between a
// live pull-config and two sequential bundle applications. Entities populated by the pull-config
// must never be affected by a bundle push. Bundle-owned data must still be cleaned correctly across bundle versions.
func TestBase_NewBundle_keepsPullConfigData(t *testing.T) {
	t.Parallel()

	brpResponse := []byte(`[
  {"bsn":"111111110","gemeente":"001"},
  {"bsn":"222222220","gemeente":"002"}
]`)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(brpResponse)
	}))
	defer ts.Close()

	pullCfg := `description: "integration test pull-config"
sources:
  - name: "BRP"
    requests:
      - name: "burgers"
        method: "GET"
        uri: "` + ts.URL + `"
        timeout: "2s"
        interval: "50ms"
        initialInterval: "1ms"
        contentType: "application/json"
        mapping:
          entities:
            - typeValue: "burger"
              idField: "bsn"
              attributes:
                - base: ""
                  map:
                    - keyValue: "gemeente"
                      valueField: "gemeente"
`

	dir := t.TempDir()
	path := filepath.Join(dir, "pull-config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(pullCfg), 0o600))

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ip, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithPullConfigs(path))
	require.NoError(t, err)

	// wait for the pull-config to populate the first burger, proving the real HTTP+scheduler path works.
	require.Eventually(t, func() bool {
		e, _, _ := ip.GetEntity("burger::111111110")
		return e != nil
	}, pullPollTimeout, pullPollInterval, "pull-config never populated pulled entities")

	ap, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	m := &Base{Ctx: ctx, Logger: logger, PAP: ap, PIP: ip, AuthMutex: &sync.RWMutex{}}

	// first bundle: brings in a consumer entity/attribute unrelated to the pulled BRP data.
	_, err = m.NewBundle(&bundles.Bundle{
		Version: 1,
		Attributes: map[string]*attributes.Attribute{
			"gemeente-vlierdam.enabled": {Key: "gemeente-vlierdam.enabled", Type: "xsd:boolean", Value: true},
		},
		Entities: map[string]*attributes.Entity{
			"c1": {Type: "consumer", Id: "gemeente-vlierdam"},
		},
	})
	require.NoError(t, err)

	burger1, _, err := ip.GetEntity("burger::111111110")
	require.NoError(t, err)
	assert.NotNil(t, burger1, "pull-owned entity must survive the first bundle apply")

	burger2, _, err := ip.GetEntity("burger::222222220")
	require.NoError(t, err)
	assert.NotNil(t, burger2, "pull-owned entity must survive the first bundle apply")

	consumer1, _, err := ip.GetEntity("consumer::gemeente-vlierdam")
	require.NoError(t, err)
	assert.NotNil(t, consumer1, "bundle-owned entity must be added")

	// second bundle: drops the first consumer, adds a different one - the bundle-owned data
	// must be pruned/added accordingly, but the pulled BRP data must remain untouched throughout.
	_, err = m.NewBundle(&bundles.Bundle{
		Version: 2,
		Entities: map[string]*attributes.Entity{
			"c2": {Type: "consumer", Id: "gemeente-drachtstad"},
		},
	})
	require.NoError(t, err)

	burger1, _, err = ip.GetEntity("burger::111111110")
	require.NoError(t, err)
	assert.NotNil(t, burger1, "pull-owned entity must survive the second bundle apply")

	burger2, _, err = ip.GetEntity("burger::222222220")
	require.NoError(t, err)
	assert.NotNil(t, burger2, "pull-owned entity must survive the second bundle apply")

	gone, _, err := ip.GetEntity("consumer::gemeente-vlierdam")
	require.NoError(t, err)
	assert.Nil(t, gone, "the first bundle's consumer must be pruned once it drops out of the latest bundle")

	consumer2, _, err := ip.GetEntity("consumer::gemeente-drachtstad")
	require.NoError(t, err)
	assert.NotNil(t, consumer2, "the second bundle's consumer must be added")

	oldEnabled, _, err := ip.GetAttribute("gemeente-vlierdam.enabled")
	require.NoError(t, err)
	assert.Nil(t, oldEnabled, "the first bundle's attribute must be pruned once it drops out of the latest bundle")

	// the pull-config keeps running on its own schedule the whole time; confirm it's still alive
	// and hasn't been disrupted by either bundle apply.
	require.Eventually(t, func() bool {
		e, _, _ := ip.GetEntity("burger::111111110")
		return e != nil
	}, pullPollTimeout, pullPollInterval, "pull-config stopped updating after repeated bundle applies")
}
