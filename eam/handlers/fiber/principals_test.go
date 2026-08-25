package fiber

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	settingsOAS "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

type storeStub struct {
	records map[string]principals.Record
	err     error
	calls   int
	asked   []string
}

func (s *storeStub) Upsert(context.Context, *identity.Principal) error { return nil }

func (s *storeStub) Resolve(_ context.Context, ids []string) (map[string]principals.Record, error) {
	s.calls++
	s.asked = append(s.asked, ids...)

	if s.err != nil {
		return nil, s.err
	}

	return s.records, nil
}

func newResolver(store principals.Store) principalResolver {
	return newPrincipalResolver(slog.New(slog2.NewDummyHandler(slog.LevelInfo)), []HandlerOption{WithPrincipals(store)})
}

// respond runs v through a real fiber request so the reflection walk sees what a handler sends.
func respond(t *testing.T, r principalResolver, v any) []byte {
	t.Helper()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return r.respond(c, v) })

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return body
}

// respondJSON is respond for the object-shaped responses, decoded for field assertions.
func respondJSON(t *testing.T, r principalResolver, v any) map[string]any {
	t.Helper()

	var out map[string]any
	require.NoError(t, json.Unmarshal(respond(t, r, v), &out))

	return out
}

func TestPrincipalResolver_FillsKnownAndFallsBackOnUnknown(t *testing.T) {
	t.Parallel()

	store := &storeStub{records: map[string]principals.Record{
		"sub-123": {ID: "sub-123", Kind: principals.KindUser, Name: "Ton de Vries"},
	}}

	pol := &oas.Policy{Audit: oas.ObjectAudit{
		CreatedBy: oas.Principal{Id: "sub-123"},
		UpdatedBy: &oas.Principal{Id: "sub-unseen"},
	}}

	out := respondJSON(t, newResolver(store), pol)
	audit := out["audit"].(map[string]any)

	created := audit["createdBy"].(map[string]any)
	assert.Equal(t, "Ton de Vries", created["name"])
	assert.Equal(t, "user", created["kind"])

	// An id the manager has never seen keeps the raw id and gains no name, so the client shows
	// the id rather than nothing. This is the only path available for *_audit rows, which carry
	// no foreign key to principal.
	updated := audit["updatedBy"].(map[string]any)
	assert.Equal(t, "sub-unseen", updated["id"])
	assert.NotContains(t, updated, "name")
}

func TestPrincipalResolver_BatchesEveryReferenceInOneLookup(t *testing.T) {
	t.Parallel()

	store := &storeStub{records: map[string]principals.Record{}}

	list := []*oas.Policy{
		{Audit: oas.ObjectAudit{CreatedBy: oas.Principal{Id: "a"}}},
		{Audit: oas.ObjectAudit{CreatedBy: oas.Principal{Id: "b"}}},
		{Audit: oas.ObjectAudit{CreatedBy: oas.Principal{Id: "c"}}},
	}

	respond(t, newResolver(store), list)

	// A list view must cost one lookup, not one per row.
	assert.Equal(t, 1, store.calls)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, store.asked)
}

func TestPrincipalResolver_ResolveFailureStillResponds(t *testing.T) {
	t.Parallel()

	store := &storeStub{err: errors.New("principal store unavailable")}
	pol := &oas.Policy{Audit: oas.ObjectAudit{CreatedBy: oas.Principal{Id: "sub-123"}}}

	// Names are display data: a read must not fail because the principal store is down.
	out := respondJSON(t, newResolver(store), pol)
	assert.Equal(t, "sub-123", out["audit"].(map[string]any)["createdBy"].(map[string]any)["id"])
}

func TestPrincipalResolver_NoStoreLeavesIdsUntouched(t *testing.T) {
	t.Parallel()

	r := newPrincipalResolver(slog.New(slog2.NewDummyHandler(slog.LevelInfo)), nil)
	pol := &oas.Policy{Audit: oas.ObjectAudit{CreatedBy: oas.Principal{Id: "sub-123"}}}

	out := respondJSON(t, r, pol)
	created := out["audit"].(map[string]any)["createdBy"].(map[string]any)
	assert.Equal(t, "sub-123", created["id"])
	assert.NotContains(t, created, "name")
}

// TestHandlers_NeverSerialiseWithoutResolving guards the design claim behind respond: that a
// response carrying attribution cannot forget to resolve it.
//
// The reflection walk makes the *shape* of a response irrelevant, but it still has to be reached.
// A chained `req.Status(...).JSON(v)` silently bypassed it once already, so this asserts the
// property directly at the call sites rather than endpoint by endpoint.
func TestHandlers_NeverSerialiseWithoutResolving(t *testing.T) {
	t.Parallel()

	// Handlers whose responses carry a Principal. The AuthZEN and decision-log handlers do not.
	files := []string{
		"policies.go", "attributes.go", "entities.go",
		"tags.go", "settings.go", "bundles.go", "languages.go",
	}

	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			src, err := os.ReadFile(name)
			require.NoError(t, err)

			for i, line := range strings.Split(string(src), "\n") {
				if strings.Contains(line, ".JSON(") {
					assert.Contains(t, line, "h.respond(",
						"%s:%d serializes without resolving principals; use h.respond(req, v)", name, i+1)
				}
			}
		})
	}
}

func TestMayContainPrincipal(t *testing.T) {
	t.Parallel()

	// The walk must not descend into a logo or an arbitrary attribute payload: those are the
	// large values on the response, and no element of them can ever be a Principal.
	assert.False(t, mayContainPrincipal(reflect.TypeOf([]byte(nil))), "[]byte is a dead end")
	assert.False(t, mayContainPrincipal(reflect.TypeOf("")), "string is a dead end")
	assert.False(t, mayContainPrincipal(reflect.TypeOf([]string(nil))))

	assert.True(t, mayContainPrincipal(reflect.TypeOf(oas.Principal{})))
	assert.True(t, mayContainPrincipal(reflect.TypeOf([]*oas.Policy(nil))))
	assert.True(t, mayContainPrincipal(reflect.TypeOf(map[string]oas.Policy(nil))))
	// An `any` field could hold anything, so it must still be followed.
	assert.True(t, mayContainPrincipal(reflect.TypeOf((*any)(nil)).Elem()))
}

func TestPrincipalResolver_SkipsLargeByteSlices(t *testing.T) {
	t.Parallel()

	store := &storeStub{records: map[string]principals.Record{
		"sub-1": {ID: "sub-1", Kind: principals.KindUser, Name: "Ton"},
	}}

	settings := &settingsOAS.Settings{
		HeaderTitle: "ACME",
		Logo:        make([]byte, 256*1024),
		UpdatedBy:   &settingsOAS.Principal{Id: "sub-1"},
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return newResolver(store).respond(c, settings) })

	req, err := http.NewRequest(http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	// The principal is still resolved; the 256 KB logo beside it is simply not walked.
	assert.Equal(t, "Ton", settings.UpdatedBy.Name)
}
