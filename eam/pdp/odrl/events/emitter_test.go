package events

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/export"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/cloudevents"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

const regoPath = "../../../../testdata/policies/opa/dienst_toeslagen/zorgtoeslag.rego"

func newPAP(t *testing.T) pap2.PAP {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return pap2.New(context.Background(), logger, pap2.WithLanguage("rego"))
}

func addRego(t *testing.T, p pap2.PAP) {
	t.Helper()

	content, err := os.ReadFile(regoPath)
	require.NoError(t, err)

	pol, err2 := pap2.NewPolicyFromData("dienst_toeslagen/zorgtoeslag", "opa", "", "", bytes.NewReader(content))
	require.NoError(t, err2)
	_, err3 := p.Create(pol)
	require.NoError(t, err3)
}

func TestEmitterPushesCloudEvent(t *testing.T) {
	t.Parallel()

	type received struct {
		contentType string
		body        []byte
	}
	got := make(chan received, 4)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got <- received{contentType: r.Header.Get("Content-Type"), body: body}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	p := newPAP(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exp := export.New(p, nil, "https://pap.example.gov.nl")

	emitter := New(context.Background(), logger, p, exp,
		[]string{srv.URL}, "", "urn:test:pap", WithClient(srv.Client()))
	p.AddEventSink(emitter)

	addRego(t, p) // Create triggers PolicyAdded → emitter.
	emitter.Wait()

	select {
	case r := <-got:
		assert.Equal(t, cloudevents.ContentTypeStructured, r.contentType)

		event, err := cloudevents.ParseStructured(r.body)
		require.NoError(t, err)
		assert.Equal(t, EventTypeUpdated, event.Type)
		assert.Equal(t, "urn:test:pap", event.Source)
		assert.Equal(t, "opa/dienst_toeslagen/zorgtoeslag", event.Subject)
		assert.Equal(t, mime.MimeTypeTurtle, event.DataContentType)
		assert.NotEmpty(t, event.DataVersion(), "dataversion extension must carry the payload hash")

		doc, err2 := model.Parse(bytes.NewReader(event.Data), mime.MimeTypeTurtle)
		require.NoError(t, err2)
		require.Len(t, doc.Artifacts, 1)
		assert.Equal(t, "data.doelbinding.zorgtoeslag", doc.Artifacts[0].Entrypoint)
	case <-time.After(5 * time.Second):
		t.Fatal("no event received")
	}
}

func TestEmitterRetriesWithBackoff(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	p := newPAP(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exp := export.New(p, nil, "")

	emitter := New(context.Background(), logger, p, exp,
		[]string{srv.URL}, "", "", WithClient(srv.Client()), WithRetry(5, time.Millisecond))
	p.AddEventSink(emitter)

	addRego(t, p)
	emitter.Wait()

	assert.Equal(t, int32(3), calls.Load(), "delivery must be retried until it succeeds (at-least-once)")
}

func TestEmitterFileExport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	p := newPAP(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exp := export.New(p, nil, "https://pap.example.gov.nl")

	emitter := New(context.Background(), logger, p, exp, nil, dir, "")
	p.AddEventSink(emitter)

	addRego(t, p)
	emitter.Wait()

	path := filepath.Join(dir, "opa", "dienst_toeslagen", "zorgtoeslag.ttl")
	data, err := os.ReadFile(path)
	require.NoError(t, err, "policy change must be exported as a Turtle file")

	doc, err2 := model.Parse(bytes.NewReader(data), mime.MimeTypeTurtle)
	require.NoError(t, err2)
	assert.Len(t, doc.Artifacts, 1)

	// removal deletes the file again.
	prev, lastIndex, err3 := p.Read("opa", "dienst_toeslagen/zorgtoeslag")
	require.NoError(t, err3)
	_, err4 := p.Delete(prev, lastIndex)
	require.NoError(t, err4)
	emitter.Wait()

	_, err5 := os.Stat(path)
	assert.True(t, os.IsNotExist(err5), "removed policy must remove the export file")
}

func TestEmitterIgnoresNonPolicyEvents(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	emitter := New(context.Background(), logger, p, export.New(p, nil, ""), nil, "", "")

	emitter.Handle(models.AttributeAdded, "some/key") // must not panic or emit.
	emitter.Wait()
}
