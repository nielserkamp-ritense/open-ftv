package fiber

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// stubSettings is a test double for settings.SettingsPersister.
type stubSettings struct {
	getErr error
	putErr error
	last   *oas.Settings
}

func (s *stubSettings) GetSettings(context.Context) (*oas.Settings, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}

	return &oas.Settings{
		HeaderTitle: "OpenFTV beheeromgeving",
		HeaderColor: "#F7E8E8",
		TitleColor:  "#000000",
	}, nil
}

func (s *stubSettings) UpdateSettings(_ context.Context, _ identity.Principal, in *oas.Settings) (*oas.Settings, error) {
	if s.putErr != nil {
		return nil, s.putErr
	}

	out := *in
	s.last = &out

	return &out, nil
}

type denyAuthorizer struct{}

func (denyAuthorizer) Authorize(*authorization.Request) (*models.Response, identity.Principal, error) {
	return &models.Response{Allowed: false}, identity.Principal{}, nil
}

func (denyAuthorizer) Identify(*authorization.Request) (*authorization.RequestPrincipal, error) {
	return authorization.SystemRequestPrincipal(), nil
}

func (denyAuthorizer) Decide(context.Context, *authorization.RequestPrincipal, string, *models.Entity) error {
	return authorization.ErrForbidden
}

func settingsAuth(t *testing.T) authorization.Authorizer {
	t.Helper()

	auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
	require.NotNil(t, auth)

	return auth
}

func settingsLogger() *slog.Logger {
	return slog.New(slog2.NewDummyHandler(slog.LevelDebug))
}

func validSettingsJSON(mods ...func(map[string]any)) []byte {
	m := map[string]any{
		"headerTitle": "ACME",
		"headerColor": "#112233",
		"titleColor":  "#AABBCC",
	}
	for _, mod := range mods {
		mod(m)
	}

	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return b
}

func TestNewSettingsHandler(t *testing.T) {
	t.Parallel()

	sh := NewSettingsHandler(settingsLogger(), &stubSettings{}, nil)
	require.NotNil(t, sh)
	assert.NotNil(t, sh.logger)
	assert.NotNil(t, sh.store)
}

func TestSettingsHandler_GetSettings(t *testing.T) {
	t.Parallel()

	sh := NewSettingsHandler(settingsLogger(), &stubSettings{}, settingsAuth(t))
	srv := fiber.New()
	srv.Get("/v1/settings", sh.GetSettings)

	resp, err := srv.Test(httptest.NewRequest(fiber.MethodGet, "/v1/settings", http.NoBody), 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, SettingsVersion, resp.Header.Get(HeaderVersion))
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var got oas.Settings
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, "OpenFTV beheeromgeving", got.HeaderTitle)
	assert.Equal(t, "#F7E8E8", got.HeaderColor)
	assert.Equal(t, "#000000", got.TitleColor)
}

func TestSettingsHandler_GetSettings_StoreError(t *testing.T) {
	t.Parallel()

	sh := NewSettingsHandler(settingsLogger(), &stubSettings{getErr: errors.New("db down")}, settingsAuth(t))
	srv := fiber.New()
	srv.Get("/v1/settings", sh.GetSettings)

	resp, err := srv.Test(httptest.NewRequest(fiber.MethodGet, "/v1/settings", http.NoBody), 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestSettingsHandler_GetSettings_Forbidden(t *testing.T) {
	t.Parallel()

	logger := settingsLogger()
	sh := NewSettingsHandler(logger, &stubSettings{}, denyAuthorizer{})
	srv := fiber.New(fiber.Config{ErrorHandler: server.ErrorHandler(logger)})
	srv.Get("/v1/settings", sh.GetSettings)

	resp, err := srv.Test(httptest.NewRequest(fiber.MethodGet, "/v1/settings", http.NoBody), 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSettingsHandler_PutSettings(t *testing.T) {
	t.Parallel()

	store := &stubSettings{}
	sh := NewSettingsHandler(settingsLogger(), store, settingsAuth(t))
	srv := fiber.New()
	srv.Put("/v1/settings", sh.PutSettings)

	body := validSettingsJSON(func(m map[string]any) {
		m["logo"] = base64.StdEncoding.EncodeToString([]byte{0x89, 0x50, 0x4e, 0x47})
		m["logoMediaType"] = "image/png"
	})
	req := httptest.NewRequest(fiber.MethodPut, "/v1/settings", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := srv.Test(req, 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, SettingsVersion, resp.Header.Get(HeaderVersion))
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NotNil(t, store.last)
	assert.Equal(t, "ACME", store.last.HeaderTitle)
	assert.Equal(t, oas.Imagepng, store.last.LogoMediaType)
	assert.NotEmpty(t, store.last.Logo)
}

func TestSettingsHandler_PutSettings_ClearsLogoMediaTypeWhenLogoEmpty(t *testing.T) {
	t.Parallel()

	store := &stubSettings{}
	sh := NewSettingsHandler(settingsLogger(), store, settingsAuth(t))
	srv := fiber.New()
	srv.Put("/v1/settings", sh.PutSettings)

	body := validSettingsJSON(func(m map[string]any) {
		m["logoMediaType"] = "image/png" // stale type must be cleared when logo is absent
	})
	req := httptest.NewRequest(fiber.MethodPut, "/v1/settings", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := srv.Test(req, 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NotNil(t, store.last)
	assert.Empty(t, store.last.LogoMediaType)
}

func TestSettingsHandler_PutSettings_StoreError(t *testing.T) {
	t.Parallel()

	sh := NewSettingsHandler(settingsLogger(), &stubSettings{putErr: errors.New("write failed")}, settingsAuth(t))
	srv := fiber.New()
	srv.Put("/v1/settings", sh.PutSettings)

	req := httptest.NewRequest(fiber.MethodPut, "/v1/settings", bytes.NewReader(validSettingsJSON()))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := srv.Test(req, 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestSettingsHandler_PutSettings_Forbidden(t *testing.T) {
	t.Parallel()

	logger := settingsLogger()
	sh := NewSettingsHandler(logger, &stubSettings{}, denyAuthorizer{})
	srv := fiber.New(fiber.Config{ErrorHandler: server.ErrorHandler(logger)})
	srv.Put("/v1/settings", sh.PutSettings)

	req := httptest.NewRequest(fiber.MethodPut, "/v1/settings", bytes.NewReader(validSettingsJSON()))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := srv.Test(req, 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSettingsHandler_PutSettings_InvalidJSON(t *testing.T) {
	t.Parallel()

	sh := NewSettingsHandler(settingsLogger(), &stubSettings{}, settingsAuth(t))
	srv := fiber.New()
	srv.Put("/v1/settings", sh.PutSettings)

	req := httptest.NewRequest(fiber.MethodPut, "/v1/settings", bytes.NewReader([]byte(`{not-json`)))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := srv.Test(req, 60000)
	require.NoError(t, err)

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSettingsHandler_PutSettings_Validation(t *testing.T) {
	t.Parallel()

	sh := NewSettingsHandler(settingsLogger(), &stubSettings{}, settingsAuth(t))
	srv := fiber.New()
	srv.Put("/v1/settings", sh.PutSettings)

	tests := []struct {
		name     string
		body     []byte
		want     string
		wantCode string
	}{
		{
			name:     "empty title and bad header color",
			body:     []byte(`{"headerTitle":"","headerColor":"not-a-color","titleColor":"#000000"}`),
			want:     "header title must be filled",
			wantCode: "E08005",
		},
		{
			name:     "title too long",
			body:     validSettingsJSON(func(m map[string]any) { m["headerTitle"] = strings.Repeat("x", 201) }),
			want:     "header title too long",
			wantCode: "E08010",
		},
		{
			name:     "empty header color",
			body:     validSettingsJSON(func(m map[string]any) { m["headerColor"] = "" }),
			want:     "header color must be filled",
			wantCode: "E08015",
		},
		{
			name:     "bad header color",
			body:     validSettingsJSON(func(m map[string]any) { m["headerColor"] = "#xyz" }),
			want:     "header color must be a hex color code",
			wantCode: "E08020",
		},
		{
			name:     "empty title color",
			body:     validSettingsJSON(func(m map[string]any) { m["titleColor"] = "" }),
			want:     "title color must be filled",
			wantCode: "E08025",
		},
		{
			name:     "bad title color",
			body:     validSettingsJSON(func(m map[string]any) { m["titleColor"] = "#xyz" }),
			want:     "title color must be a hex color code",
			wantCode: "E08030",
		},
		{
			name: "logo without valid media type",
			body: validSettingsJSON(func(m map[string]any) {
				m["logo"] = base64.StdEncoding.EncodeToString([]byte("png"))
				m["logoMediaType"] = "image/gif"
			}),
			want:     "logo media type must be one of",
			wantCode: "E08035",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(fiber.MethodPut, "/v1/settings", bytes.NewReader(tc.body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err := srv.Test(req, 60000)
			require.NoError(t, err)

			require.NotNil(t, resp)
			defer resp.Body.Close()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Contains(t, string(b), tc.want, fmt.Sprintf("body=%s", b))
			assert.Contains(t, string(b), tc.wantCode, fmt.Sprintf("body=%s", b))
		})
	}
}
