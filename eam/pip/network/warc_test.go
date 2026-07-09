package network

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

func TestManager_Execute_WARCAndRecorder(t *testing.T) {
	t.Parallel()

	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"leeftijd": 42, "woonplaats": "Den Haag"}`))
	}))
	defer svc.Close()

	var decData = `
attributes:
  - base: ""
    map:
      - keyValue: leeftijd
        valueField: leeftijd
        typeValue: xsd:short
      - keyValue: woonplaats
        valueField: woonplaats
        typeValue: xsd:string
`
	var dec ResponseMapping
	require.NoError(t, yaml.Unmarshal([]byte(decData), &dec))

	dir := t.TempDir()
	w, err := warc.NewWriter(warc.Config{Dir: dir})
	require.NoError(t, err)
	defer w.Close()

	type recorded struct{ kind, key, traceID, spanID, warcFile string }
	var mu sync.Mutex
	var refs []recorded

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	m := &manager{
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
		attributes:    models.NewAttributeSet(),
		entities:      models.NewEntitySet(),
		newAttributes: models.NewAttributeSet,
		warc:          w,
		recorder: func(kind, key, traceID, spanID, warcFile, _ string) {
			mu.Lock()
			refs = append(refs, recorded{kind, key, traceID, spanID, warcFile})
			mu.Unlock()
		},
	}

	req := &Request{
		Name:    "persoon",
		Method:  http.MethodGet,
		URI:     svc.URL + "/v1/persoon",
		Timeout: 5 * time.Second,
		Mapping: &dec,
	}

	require.NoError(t, m.execute(logger, req))

	// the decoded attributes must be present.
	assert.NotNil(t, m.attributes.GetAttribute("leeftijd"))
	assert.NotNil(t, m.attributes.GetAttribute("woonplaats"))

	// both attributes must have a recorded source reference with the same trace/span/file.
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, refs, 2)
	assert.Equal(t, "attribute", refs[0].kind)
	assert.Equal(t, refs[0].spanID, refs[1].spanID)
	assert.Len(t, refs[0].traceID, 32)
	assert.Len(t, refs[0].spanID, 16)
	require.NotEmpty(t, refs[0].warcFile)

	// the WARC file must contain a request/response pair with matching identifiers.
	data, err := os.ReadFile(filepath.Join(dir, refs[0].warcFile))
	require.NoError(t, err)

	recs, err := warc.Read(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, recs, 2)

	assert.Equal(t, warc.TypeRequest, recs[0].Type)
	assert.Equal(t, warc.TypeResponse, recs[1].Type)
	assert.Equal(t, recs[0].RecordID, recs[1].ConcurrentTo)
	assert.Equal(t, recs[1].RecordID, recs[0].ConcurrentTo)
	assert.Equal(t, refs[0].traceID, recs[0].Custom[warc.HeaderTraceID])
	assert.Equal(t, refs[0].spanID, recs[1].Custom[warc.HeaderSpanID])
	assert.Contains(t, string(recs[0].Block), "GET /v1/persoon HTTP/1.1")
	assert.Contains(t, string(recs[1].Block), "Den Haag")
}
