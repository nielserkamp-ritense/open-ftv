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

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/opa-embedded"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewAttributesHandler(t *testing.T) {
	t.Parallel()

	t.Run("test new attributes handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/policies/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ah := NewAttributesHandler(logger, controller.PIP(), nil)
		require.NotNil(t, ah)
	})
}

func TestAttributesHandler_GetAttributes(t *testing.T) {
	t.Parallel()

	t.Run("get attributes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("rego"), pap2.WithFileStore("../../../testdata/policies/opa", true))
		require.NotNil(t, p2)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ah := NewAttributesHandler(logger, controller.PIP(), nil)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/attributes", ah.GetAttributes)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/attributes", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, AttributesVersion, resp.Header.Get(HeaderVersion))

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
	t.Parallel()

	t.Run("get attributes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/non_existing_folder", true))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("rego"), pap2.WithFileStore("../../../testdata/policies/opa", true))
		require.NotNil(t, p2)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ah := NewAttributesHandler(logger, controller.PIP(), nil)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/attributes", ah.GetAttributes)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/attributes", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, AttributesVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestAttributesHandler_GetAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		key        string
		wantStatus int
		wantVer    string
	}{
		{name: "no ID", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "bad ID", key: "qqq99", wantStatus: fiber.StatusNotFound, wantVer: AttributesVersion},
		{name: "good ID", key: "werktijden", wantStatus: fiber.StatusOK, wantVer: AttributesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
			require.NotNil(t, p1)

			p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/policies/cedar", true))
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller.PIP(), nil)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Get("/v1/attribute/:key", ah.GetAttribute)

			req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/attribute/"+tc.key, nil)
			resp, err2 := srv.Test(req, 100)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
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
	t.Parallel()

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
		wantVer    string
	}{
		{name: "no ID", body: bytes.NewBufferString(""), timeout: 100 * time.Millisecond, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 100 * time.Millisecond, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "no body", key: "xyz", timeout: 100 * time.Millisecond, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "bad body", key: "xyz", body: bytes.NewBufferString("not a json payload"), timeout: 100 * time.Millisecond, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "duplicate key", key: "werktijden", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusConflict, wantVer: AttributesVersion},
		{name: "mismatched keys", key: "xyz", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "all good", key: "key", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusOK, wantVer: AttributesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
			require.NotNil(t, p1)

			p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/policies/cedar", true))
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller.PIP(), nil)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Put("/v1/attribute/:key", ah.PutAttribute)

			req := httptest.NewRequest(fiber.MethodPut, "/v1/attribute/"+tc.key, tc.body)
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
	t.Parallel()

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
		wantVer    string
	}{
		{name: "no ID", body: bytes.NewBufferString(""), timeout: 100 * time.Millisecond, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 100 * time.Millisecond, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "no body", key: "xyz", timeout: 100 * time.Millisecond, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "bad body", key: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: 100 * time.Millisecond, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "mismatched keys", key: "xyz", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "not found", key: "key", body: bytes.NewBuffer(body1), timeout: 5 * time.Second, wantStatus: fiber.StatusNotFound, wantVer: AttributesVersion},
		{name: "all good", key: "werktijden", body: bytes.NewBuffer(body2), timeout: 5 * time.Second, wantStatus: fiber.StatusOK, wantVer: AttributesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
			require.NotNil(t, p1)

			p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/policies/cedar", true))
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller.PIP(), nil)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Post("/v1/attribute/:key", ah.PostAttribute)

			req := httptest.NewRequest(fiber.MethodPost, "/v1/attribute/"+tc.key, tc.body)
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
	t.Parallel()

	testCases := []struct {
		name       string
		key        string
		wantStatus int
		wantVer    string
	}{
		{name: "no ID", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", key: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest, wantVer: AttributesVersion},
		{name: "not found", key: "xyz", wantStatus: fiber.StatusNotFound, wantVer: AttributesVersion},
		{name: "all good", key: "werktijden", wantStatus: fiber.StatusOK, wantVer: AttributesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
			require.NotNil(t, p1)

			p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/policies/cedar", true))
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ah := NewAttributesHandler(logger, controller.PIP(), nil)
			require.NotNil(t, ah)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Delete("/v1/attribute/:key", ah.DeleteAttribute)

			req := httptest.NewRequest(fiber.MethodDelete, "/v1/attribute/"+tc.key, nil)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, 100)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
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
