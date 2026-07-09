package server

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

func newEventsApp(t *testing.T) (*fiber.App, pip2.PIP, string) {
	t.Helper()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	p := pip2.New(context.Background(), logger)

	dir := t.TempDir()
	w, err := warc.NewWriter(warc.Config{Dir: dir})
	require.NoError(t, err)
	t.Cleanup(func() { _ = w.Close() })

	h := newEventsHandler(logger, p, nil, w)

	app := fiber.New()
	app.Post("/v1/events", h.PostEvent)
	app.Get("/v1/sourcerefs", h.GetSourceRefs)
	app.Get("/v1/sourceref", h.GetSourceRef)

	return app, p, dir
}

func TestPostEvent_Structured(t *testing.T) {
	t.Parallel()

	app, p, dir := newEventsApp(t)

	body := `{
		"specversion": "1.0",
		"id": "evt-100",
		"source": "https://brp.example",
		"type": "nl.ftv.pip.update",
		"dataversion": "3.1.4",
		"sequence": "55",
		"data": {
			"attributes": [{"key": "leeftijd", "value": 42, "type": "xsd:short"}],
			"entities": [{"type": "service", "id": "brp-personen"}]
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	assert.Equal(t, "accepted", out["status"])
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", out["traceId"], "inbound trace id must be preserved")

	// the attribute and entity must be stored.
	attr := p.GetAttribute("leeftijd")
	require.NotNil(t, attr)
	assert.Equal(t, int64(42), attr.Value())
	require.NotNil(t, p.GetEntity(models.EntityUID("service", "brp-personen")))

	// source references must carry the version tags and WARC file.
	refs, ok := p.(pip2.SourceReferencer)
	require.True(t, ok)

	ref, found := refs.SourceRef("leeftijd")
	require.True(t, found)
	assert.Equal(t, "attribute", ref.Kind)
	assert.Equal(t, "3.1.4", ref.Version)
	assert.Equal(t, "55", ref.Sequence)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", ref.TraceID)
	assert.Equal(t, "00f067aa0ba902b7", ref.ParentSpanID)
	assert.NotEmpty(t, ref.SpanID)
	assert.NotEmpty(t, ref.WARCFile)

	eref, found := refs.SourceRef(models.EntityUID("service", "brp-personen"))
	require.True(t, found)
	assert.Equal(t, "entity", eref.Kind)
	assert.Equal(t, ref.SpanID, eref.SpanID, "all updates from one event share the span")

	// the WARC file must contain a request + resource record pair with the trace ids.
	data, err := os.ReadFile(filepath.Join(dir, ref.WARCFile))
	require.NoError(t, err)

	recs, err := warc.Read(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, recs, 2)
	assert.Equal(t, warc.TypeRequest, recs[0].Type)
	assert.Equal(t, warc.TypeResource, recs[1].Type)
	assert.Equal(t, ref.TraceID, recs[0].Custom[warc.HeaderTraceID])
	assert.Equal(t, ref.SpanID, recs[0].Custom[warc.HeaderSpanID])
	assert.Equal(t, "3.1.4", recs[1].Custom[warc.HeaderVersion])
	assert.Contains(t, string(recs[0].Block), "POST /v1/events")
	assert.Contains(t, string(recs[1].Block), "evt-100")
}

func TestPostEvent_Binary(t *testing.T) {
	t.Parallel()

	app, p, _ := newEventsApp(t)

	body := `{"attributes": [{"key": "woonplaats", "value": "Den Haag"}]}`

	req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ce-specversion", "1.0")
	req.Header.Set("ce-id", "evt-200")
	req.Header.Set("ce-source", "urn:example:rdw")
	req.Header.Set("ce-type", "nl.ftv.pip.update")
	req.Header.Set("ce-dataversion", "v7")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	attr := p.GetAttribute("woonplaats")
	require.NotNil(t, attr)
	assert.Equal(t, "Den Haag", attr.Value())

	refs := p.(pip2.SourceReferencer)
	ref, found := refs.SourceRef("woonplaats")
	require.True(t, found)
	assert.Equal(t, "v7", ref.Version)
}

func TestPostEvent_Idempotent(t *testing.T) {
	t.Parallel()

	app, p, _ := newEventsApp(t)

	send := func(value string) map[string]any {
		body := `{
			"specversion": "1.0", "id": "evt-300", "source": "urn:example:brp", "type": "t",
			"data": {"attributes": [{"key": "naam", "value": "` + value + `"}]}
		}`
		req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/cloudevents+json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var out map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		return out
	}

	out1 := send("eerste")
	assert.Equal(t, "accepted", out1["status"])

	// same (source, id): must be acknowledged but not reapplied.
	out2 := send("tweede")
	assert.Equal(t, "duplicate", out2["status"])

	attr := p.GetAttribute("naam")
	require.NotNil(t, attr)
	assert.Equal(t, "eerste", attr.Value(), "duplicate event must not overwrite data")
}

func TestPostEvent_Validation(t *testing.T) {
	t.Parallel()

	app, _, _ := newEventsApp(t)

	for name, tc := range map[string]struct {
		contentType string
		body        string
	}{
		"missing id":     {"application/cloudevents+json", `{"specversion":"1.0","source":"s","type":"t","data":{}}`},
		"missing source": {"application/cloudevents+json", `{"specversion":"1.0","id":"i","type":"t","data":{}}`},
		"bad envelope":   {"application/cloudevents+json", `{`},
		"no ce headers":  {"application/json", `{}`},
		"empty payload":  {"application/cloudevents+json", `{"specversion":"1.0","id":"i","source":"s","type":"t","data":{}}`},
	} {
		req := httptest.NewRequest(http.MethodPost, "/v1/events", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", tc.contentType)

		resp, err := app.Test(req)
		require.NoError(t, err, name)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, name)
	}
}

func TestSourceRefLookupEndpoints(t *testing.T) {
	t.Parallel()

	app, p, _ := newEventsApp(t)

	refs := p.(pip2.SourceReferencer)
	refs.RecordSourceRef(pip2.SourceRef{Kind: "attribute", Key: "leeftijd", SpanID: "00f067aa0ba902b7", WARCFile: "x.warc", Version: "1"})

	// list.
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/v1/sourcerefs", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var list []pip2.SourceRef
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	require.Len(t, list, 1)
	assert.Equal(t, "leeftijd", list[0].Key)

	// single.
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/v1/sourceref?key=leeftijd", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var ref pip2.SourceRef
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&ref))
	assert.Equal(t, "00f067aa0ba902b7", ref.SpanID)
	assert.Equal(t, "x.warc", ref.WARCFile)

	// missing key.
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/v1/sourceref", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// unknown key.
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/v1/sourceref?key=onbekend", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
