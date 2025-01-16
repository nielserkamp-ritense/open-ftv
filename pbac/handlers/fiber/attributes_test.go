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

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewAttributesHandler(t *testing.T) {
	t.Run("test new attributes handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ah := NewAttributesHandler(logger, controller)
		require.NotNil(t, ah)
	})
}

func TestAttributesHandler_GetAttributes(t *testing.T) {
	t.Run("get attributes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger})
		require.NotNil(t, p)

		controller := opa.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/opa", true), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ah := NewAttributesHandler(logger, controller)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/attributes", ah.GetAttributes)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/attributes", nil)
		resp, err2 := srv.Test(req, 1)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*attributes.Attribute
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, 1, len(list))
	})
}

func TestAttributesHandler_GetAttributes_Empty(t *testing.T) {
	t.Run("get attributes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/non_existing_folder", Logger: logger})
		require.NotNil(t, p)

		controller := opa.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/opa", true), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ah := NewAttributesHandler(logger, controller)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/attributes", ah.GetAttributes)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/attributes", nil)
		resp, err2 := srv.Test(req, 1)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestAttributesHandler_GetAttribute(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		wantStatus int
	}{
		{name: "no ID", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest},
		{name: "bad ID", key: "qqq99", wantStatus: fiber.StatusNotFound},
		{name: "good ID", key: "werktijden", wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Get("/v1/attribute/:key", ah.GetAttribute)

			req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/attribute/"+tc.key, nil)
			resp, err2 := srv.Test(req, -1)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var attr attributes.Attribute
				err := json.Unmarshal(b, &attr)
				require.NoError(t, err)
				assert.NotNil(t, attr)
				assert.NotEmpty(t, attr.Key)
				assert.NotEmpty(t, attr.Value)
			}
		})
	}
}

func TestAttributesHandler_PutAttribute(t *testing.T) {
	data1 := attributes.Attribute{Key: "key", Type: "xsd:string", Value: "value"}
	body1, _ := json.Marshal(data1)

	data2 := attributes.Attribute{Key: "werktijden", Type: "xsd:string", Value: "09:00-17:00"}
	body2, _ := json.Marshal(data2)

	testCases := []struct {
		name       string
		key        string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
	}{
		{name: "no ID", body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "no body", key: "xyz", timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "bad body", key: "xyz", body: bytes.NewBufferString("not a json payload"), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "duplicate key", key: "werktijden", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusConflict},
		{name: "mismatched keys", key: "xyz", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest},
		{name: "all good", key: "key", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Put("/v1/attribute/:key", ah.PutAttribute)

			req := httptest.NewRequest(fiber.MethodPut, "/v1/attribute/"+tc.key, tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var attr attributes.Attribute
				err := json.Unmarshal(b, &attr)
				require.NoError(t, err)
				assert.NotNil(t, attr)
				assert.NotEmpty(t, attr.Key)
				assert.NotEmpty(t, attr.Value)
			}
		})
	}
}

func TestAttributesHandler_PostAttribute(t *testing.T) {
	data1 := attributes.Attribute{Key: "key", Type: "xsd:string", Value: "value"}
	body1, _ := json.Marshal(data1)

	data2 := attributes.Attribute{Key: "werktijden", Type: "xsd:string", Value: "09:00-17:00"}
	body2, _ := json.Marshal(data2)

	testCases := []struct {
		name       string
		key        string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
	}{
		{name: "no ID", body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "no body", key: "xyz", timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "bad body", key: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "mismatched keys", key: "xyz", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest},
		{name: "not found", key: "key", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "all good", key: "werktijden", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Post("/v1/attribute/:key", ah.PostAttribute)

			req := httptest.NewRequest(fiber.MethodPost, "/v1/attribute/"+tc.key, tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var attr attributes.Attribute
				err := json.Unmarshal(b, &attr)
				require.NoError(t, err)
				assert.NotNil(t, attr)
				assert.NotEmpty(t, attr.Key)
				assert.NotEmpty(t, attr.Value)
			}
		})
	}
}

func TestAttributesHandler_DeleteAttribute(t *testing.T) {
	testCases := []struct {
		name       string
		key        string
		wantStatus int
	}{
		{name: "no ID", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest},
		{name: "not found", key: "xyz", wantStatus: fiber.StatusNotFound},
		{name: "all good", key: "werktijden", wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Delete("/v1/attribute/:key", ah.DeleteAttribute)

			req := httptest.NewRequest(fiber.MethodDelete, "/v1/attribute/"+tc.key, nil)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, 1)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var attr attributes.Attribute
				err := json.Unmarshal(b, &attr)
				require.NoError(t, err)
				assert.NotNil(t, attr)
				assert.NotEmpty(t, attr.Key)
				assert.NotEmpty(t, attr.Value)
			}
		})
	}
}
