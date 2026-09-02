package fiber

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewEntitiesHandler(t *testing.T) {
	t.Parallel()

	t.Run("test new entities handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
		require.NoError(t, err)
		require.NotNil(t, p1)

		p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
		require.NoError(t, err)
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		eh := NewEntitiesHandler(logger, controller.GetPIP(), nil)
		require.NotNil(t, eh)
	})
}

func TestEntitiesHandler_GetEntities(t *testing.T) {
	t.Parallel()

	t.Run("get entities", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
		require.NoError(t, err)
		require.NotNil(t, p1)

		p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/opa", true))
		require.NoError(t, err)
		require.NotNil(t, p2)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		eh := NewEntitiesHandler(logger, controller.GetPIP(), auth)
		require.NotNil(t, eh)

		srv := fiber.New()
		srv.Get("/v1/entities", eh.GetEntities)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/entities", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, EntitiesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*attributes.Entity

		err = json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 10)

		for i := range list {
			e := list[i]
			assert.NotEmpty(t, e.Type)
			assert.NotEmpty(t, e.Id)
		}
	})
}

func TestEntitiesHandler_GetEntities_Empty(t *testing.T) {
	t.Parallel()

	t.Run("get entities", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/non_existing_folder", true))
		require.NoError(t, err)
		require.NotNil(t, p1)

		p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/opa", true))
		require.NoError(t, err)
		require.NotNil(t, p2)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		eh := NewEntitiesHandler(logger, controller.GetPIP(), auth)
		require.NotNil(t, eh)

		srv := fiber.New()
		srv.Get("/v1/entities", eh.GetEntities)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/entities", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, EntitiesVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestEntitiesHandler_GetEntity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		ns         string
		id         string
		wantStatus int
		wantVer    string
	}{
		{name: "no type, ID", wantStatus: fiber.StatusNotFound},
		{name: "no type", id: "alice", wantStatus: fiber.StatusNotFound},
		{name: "no ID", ns: "service", wantStatus: fiber.StatusNotFound},
		{name: "very long type", ns: strings.Repeat("x", 501), id: "alice", wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "very long ID", ns: "service", id: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "not found", ns: "service", id: "harry", wantStatus: fiber.StatusNotFound, wantVer: EntitiesVersion},
		{name: "found", ns: "app", id: "app1", wantStatus: fiber.StatusOK, wantVer: EntitiesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			eh := NewEntitiesHandler(logger, controller.GetPIP(), auth)
			require.NotNil(t, eh)

			srv := fiber.New()
			srv.Get("/v1/entity/:type/:id", eh.GetEntity)

			req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/entity/"+tc.ns+"/"+tc.id, nil)
			resp, err2 := srv.Test(req)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var e attributes.Entity
				err := json.Unmarshal(b, &e)
				require.NoError(t, err)
				assert.NotNil(t, e)
				assert.NotEmpty(t, e.Type)
				assert.NotEmpty(t, e.Id)
			}
		})
	}
}

func TestEntitiesHandler_PostEntity(t *testing.T) {
	t.Parallel()

	data1 := attributes.Entity{Type: "type", Id: "id", Metadata: attributes.Metadata{Title: "title"}}
	body1, _ := json.Marshal(data1)

	data2 := attributes.Entity{Type: "app", Id: "app1", Metadata: attributes.Metadata{Title: "title2"}}
	body2, _ := json.Marshal(data2)

	testCases := []struct {
		name       string
		ns         string
		id         string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
		wantVer    string
	}{
		{name: "no type", id: "id", body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "no ID", ns: "type", body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "very long type", id: "app1", ns: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "very long ID", ns: "app", id: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "no body", ns: "app", id: "xyz", timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "bad body", ns: "app", id: "xyz", body: bytes.NewBufferString("not a json payload"), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "duplicate key", ns: "app", id: "app1", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusConflict, wantVer: EntitiesVersion},
		{name: "mismatched type", ns: "app", id: "id", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "mismatched id", ns: "type", id: "xyz", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "all good", ns: "type", id: "id", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusCreated, wantVer: EntitiesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			eh := NewEntitiesHandler(logger, controller.GetPIP(), auth)
			require.NotNil(t, eh)

			srv := fiber.New()
			srv.Post("/v1/entity/:type/:id", eh.PostEntity)

			req := httptest.NewRequest(fiber.MethodPost, "/v1/entity/"+tc.ns+"/"+tc.id, tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var e attributes.Entity
				err := json.Unmarshal(b, &e)
				require.NoError(t, err)
				assert.NotNil(t, e)
				assert.NotEmpty(t, e.Type)
				assert.NotEmpty(t, e.Id)
			}
		})
	}
}

func TestEntitiesHandler_PutEntity(t *testing.T) {
	t.Parallel()

	data1 := attributes.Entity{Type: "type", Id: "id", Metadata: attributes.Metadata{Title: "title"}}
	body1, _ := json.Marshal(data1)

	data2 := attributes.Entity{Type: "app", Id: "app1", Metadata: attributes.Metadata{Title: "title2"}}
	body2, _ := json.Marshal(data2)

	testCases := []struct {
		name       string
		ns         string
		id         string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
		wantVer    string
	}{
		{name: "no type", id: "app1", body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "no ID", ns: "app", body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "very long type", id: "app1", ns: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "very long ID", ns: "app", id: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "no body", ns: "app", id: "xyz", timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "bad body", ns: "app", id: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "mismatched type", ns: "xyz", id: "app1", body: bytes.NewBuffer(body2), timeout: 60 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "mismatched is", ns: "app", id: "xyz", body: bytes.NewBuffer(body2), timeout: 60 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "not found", ns: "type", id: "id", body: bytes.NewBuffer(body1), timeout: 60 * time.Second, wantStatus: fiber.StatusNotFound, wantVer: EntitiesVersion},
		{name: "all good", ns: "app", id: "app1", body: bytes.NewBuffer(body2), timeout: 60 * time.Second, wantStatus: fiber.StatusOK, wantVer: EntitiesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			ah := NewEntitiesHandler(logger, controller.GetPIP(), auth)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Put("/v1/entity/:type/:id", ah.PutEntity)

			req := httptest.NewRequest(fiber.MethodPut, "/v1/entity/"+tc.ns+"/"+tc.id, tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var e attributes.Entity
				err := json.Unmarshal(b, &e)
				require.NoError(t, err)
				assert.NotNil(t, e)
				assert.NotEmpty(t, e.Type)
				assert.NotEmpty(t, e.Id)
			}
		})
	}
}

func TestEntitiesHandler_DeleteEntity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		ns         string
		id         string
		wantStatus int
		wantVer    string
	}{
		{name: "no type", id: "app1", wantStatus: fiber.StatusNotFound},
		{name: "no ID", ns: "app", wantStatus: fiber.StatusNotFound},
		{name: "very long type", ns: strings.Repeat("x", 501), id: "app1", wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "very long ID", ns: "app", id: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest, wantVer: EntitiesVersion},
		{name: "not found", ns: "app", id: "xyz", wantStatus: fiber.StatusNotFound, wantVer: EntitiesVersion},
		{name: "all good", ns: "app", id: "app1", wantStatus: fiber.StatusOK, wantVer: EntitiesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			ah := NewEntitiesHandler(logger, controller.GetPIP(), auth)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Delete("/v1/entity/:type/:id", ah.DeleteEntity)

			req := httptest.NewRequest(fiber.MethodDelete, "/v1/entity/"+tc.ns+"/"+tc.id, nil)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var e attributes.Entity
				err := json.Unmarshal(b, &e)
				require.NoError(t, err)
				assert.NotNil(t, e)
				assert.NotEmpty(t, e.Type)
				assert.NotEmpty(t, e.Id)
			}
		})
	}
}
