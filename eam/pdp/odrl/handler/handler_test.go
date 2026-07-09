package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/export"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/importer"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/cloudevents"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

const (
	regoPath = "../../../../testdata/policies/opa/brp/burgerzaken.rego"
	regoSHA  = "e1482e92c4cf0faa3c88bfdc54bc39912fba6d5f6e8db298bc364ac7d5e8bfa8"
)

func newApp(t *testing.T, client *http.Client) (*fiber.App, pap2.PAP) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	p := pap2.New(context.Background(), logger, pap2.WithLanguage("rego"))

	exp := export.New(p, nil, "https://pap.example.gov.nl")
	imp := importer.New(p, logger, client)

	app := fiber.New()
	h := New(logger, p, exp, imp)
	h.Register(app.Group("/v1"))
	return app, p
}

func regoServer(t *testing.T) *httptest.Server {
	t.Helper()

	content, err := os.ReadFile(regoPath)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func odrlDoc(downloadURL string) string {
	return fmt.Sprintf(`@prefix apnl: <https://standaarden.overheid.nl/odrl-ap-nl/> .
@prefix odrl: <http://www.w3.org/ns/odrl/2/> .
@prefix dct:  <http://purl.org/dc/terms/> .
@prefix dcat: <http://www.w3.org/ns/dcat#> .
@prefix ex:   <https://example.gov.nl/brp/> .

ex:artefact a apnl:RegoModule ;
    dct:format "application/vnd.rego" ;
    dcat:downloadURL <%s> ;
    apnl:entrypoint "data.doelbinding.burgerzaken.allow" ;
    apnl:sha256 "%s" .

ex:overeenkomst a odrl:Agreement ;
    dct:title "Testovereenkomst"@nl ;
    dct:publisher <https://identifier.overheid.nl/tooi/id/oorg/oorg10103> ;
    odrl:profile apnl: ;
    odrl:uid ex:overeenkomst ;
    odrl:permission [ a odrl:Permission ;
        odrl:action odrl:read ;
        odrl:constraint [ a odrl:Constraint ;
            odrl:leftOperand apnl:verwerkingsverzoek ;
            odrl:operator apnl:conformsToPolicy ;
            odrl:rightOperand ex:artefact ] ] .
`, downloadURL, regoSHA)
}

func addRego(t *testing.T, p pap2.PAP) {
	t.Helper()

	content, err := os.ReadFile(regoPath)
	require.NoError(t, err)

	pol, err2 := pap2.NewPolicyFromData("brp/burgerzaken", "opa", "", "", bytes.NewReader(content))
	require.NoError(t, err2)
	_, err3 := p.Create(pol)
	require.NoError(t, err3)
}

func TestExportEndpoint(t *testing.T) {
	t.Parallel()

	app, p := newApp(t, nil)
	addRego(t, p)

	// default: Turtle.
	req := httptest.NewRequest(http.MethodGet, "/v1/odrl/export", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), mime.MimeTypeTurtle)

	body, _ := io.ReadAll(resp.Body)
	doc, err2 := model.Parse(bytes.NewReader(body), mime.MimeTypeTurtle)
	require.NoError(t, err2)
	require.Len(t, doc.Artifacts, 1)
	assert.Equal(t, regoSHA, doc.Artifacts[0].SHA256)

	// JSON-LD on request.
	req = httptest.NewRequest(http.MethodGet, "/v1/odrl/export", nil)
	req.Header.Set(fiber.HeaderAccept, mime.MimeTypeJSONLD)
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), mime.MimeTypeJSONLD)

	body, _ = io.ReadAll(resp.Body)
	doc, err2 = model.Parse(bytes.NewReader(body), mime.MimeTypeJSONLD)
	require.NoError(t, err2)
	assert.Len(t, doc.Artifacts, 1)
}

func TestExportPolicyEndpoint(t *testing.T) {
	t.Parallel()

	app, p := newApp(t, nil)
	addRego(t, p)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/v1/odrl/export/opa/brp/burgerzaken", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	doc, err2 := model.Parse(bytes.NewReader(body), mime.MimeTypeTurtle)
	require.NoError(t, err2)
	require.Len(t, doc.Artifacts, 1)
	assert.Equal(t, "data.doelbinding.burgerzaken", doc.Artifacts[0].Entrypoint)

	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/v1/odrl/export/opa/missing", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestImportEndpoint(t *testing.T) {
	t.Parallel()

	srv := regoServer(t)
	app, p := newApp(t, srv.Client())

	req := httptest.NewRequest(http.MethodPost, "/v1/odrl/import", bytes.NewReader([]byte(odrlDoc(srv.URL))))
	req.Header.Set(fiber.HeaderContentType, mime.MimeTypeTurtle)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var res importer.Result
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.Equal(t, []string{"odrl/example.gov.nl-brp-overeenkomst"}, res.Policies)
	assert.Equal(t, []string{"rego/doelbinding/burgerzaken"}, res.Artifacts)

	_, _, err2 := p.Read("rego", "doelbinding/burgerzaken")
	assert.NoError(t, err2)
}

func TestImportEndpointByURL(t *testing.T) {
	t.Parallel()

	regoSrv := regoServer(t)
	doc := odrlDoc(regoSrv.URL)

	docSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", mime.MimeTypeTurtle)
		_, _ = w.Write([]byte(doc))
	}))
	t.Cleanup(docSrv.Close)

	app, _ := newApp(t, docSrv.Client())

	body, _ := json.Marshal(map[string]string{"url": docSrv.URL + "/doc.ttl"})
	req := httptest.NewRequest(http.MethodPost, "/v1/odrl/import", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, mime.MimeTypeJSON)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestImportEndpointRejectsBadInput(t *testing.T) {
	t.Parallel()

	app, _ := newApp(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/odrl/import", bytes.NewReader([]byte("junk")))
	req.Header.Set(fiber.HeaderContentType, "application/xml")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/v1/odrl/import", http.NoBody)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestEventsEndpointStructured(t *testing.T) {
	t.Parallel()

	srv := regoServer(t)
	app, p := newApp(t, srv.Client())

	event := map[string]any{
		"specversion":     "1.0",
		"id":              "evt-1",
		"source":          "urn:test:remote-pap",
		"type":            "nl.overheid.ftv.policy.updated",
		"subject":         "opa/brp/burgerzaken",
		"datacontenttype": mime.MimeTypeTurtle,
		"data_base64":     base64.StdEncoding.EncodeToString([]byte(odrlDoc(srv.URL))),
		"dataversion":     regoSHA,
	}
	body, _ := json.Marshal(event)

	req := httptest.NewRequest(http.MethodPost, "/v1/odrl/events", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, cloudevents.ContentTypeStructured)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var res importer.Result
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.Len(t, res.Policies, 1)
	assert.Len(t, res.Artifacts, 1)

	_, _, err2 := p.Read("odrl", "example.gov.nl-brp-overeenkomst")
	assert.NoError(t, err2)

	// same (source, id) again → deduplicated.
	req = httptest.NewRequest(http.MethodPost, "/v1/odrl/events", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, cloudevents.ContentTypeStructured)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var dup map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&dup))
	assert.Equal(t, "duplicate", dup["status"])
}

func TestEventsEndpointBinary(t *testing.T) {
	t.Parallel()

	srv := regoServer(t)
	app, _ := newApp(t, srv.Client())

	req := httptest.NewRequest(http.MethodPost, "/v1/odrl/events", bytes.NewReader([]byte(odrlDoc(srv.URL))))
	req.Header.Set(fiber.HeaderContentType, mime.MimeTypeTurtle)
	req.Header.Set("ce-specversion", "1.0")
	req.Header.Set("ce-id", "evt-binary-1")
	req.Header.Set("ce-source", "urn:test:remote-pap")
	req.Header.Set("ce-type", "nl.overheid.ftv.policy.updated")
	req.Header.Set("ce-subject", "opa/brp/burgerzaken")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var res importer.Result
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
	assert.Len(t, res.Policies, 1)
	assert.Len(t, res.Artifacts, 1)
}

func TestEventsEndpointRejectsInvalid(t *testing.T) {
	t.Parallel()

	app, _ := newApp(t, nil)

	// missing required CloudEvents attributes.
	req := httptest.NewRequest(http.MethodPost, "/v1/odrl/events", bytes.NewReader([]byte(`{"id":"x"}`)))
	req.Header.Set(fiber.HeaderContentType, cloudevents.ContentTypeStructured)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
