package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/laadpalen-applicatie/config"
)

var testAssets = fstest.MapFS{
	"index.html":    {Data: []byte("<html>app</html>")},
	"assets/app.js": {Data: []byte("console.log('app')")},
}

func newMainApp(t *testing.T, env map[string]string) *fiber.App {
	t.Helper()

	s := &Services{assets: testAssets, runtimeEnv: runtimeEnvScript(env), chk: handle.NewChecks()}

	app := fiber.New()
	s.initMainRoutes(context.Background(), app)

	return app
}

func get(t *testing.T, app *fiber.App, path string) (status int, contentType, cacheControl, body string) {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, http.NoBody))
	require.NoError(t, err)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, resp.Header.Get(fiber.HeaderContentType), resp.Header.Get(fiber.HeaderCacheControl), string(b)
}

func TestRuntimeEnv(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"VITE_API_BASE_URL":      "https://vlierdam.example.com",
		"VITE_VLIERDAM_PDP1_URL": "",
		"VITE_DOELBINDING":       `laadpaal"</script><script>alert(1)`,
	}

	status, contentType, cacheControl, body := get(t, newMainApp(t, env), runtimeEnvPath)

	assert.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, "application/javascript; charset=utf-8", contentType)
	assert.Equal(t, "no-store", cacheControl)

	// a value must not be able to close the script.
	assert.NotContains(t, body, "</script>")

	values, ok := strings.CutPrefix(body, "window.__ENV__ = ")
	require.True(t, ok, "unexpected script: %s", body)
	values, ok = strings.CutSuffix(values, ";\n")
	require.True(t, ok, "unexpected script: %s", body)

	var got map[string]string
	require.NoError(t, json.Unmarshal([]byte(values), &got))
	assert.Equal(t, env, got)
}

func TestStaticAssets(t *testing.T) {
	t.Parallel()

	app := newMainApp(t, nil)

	tests := []struct {
		name        string
		path        string
		contentType string
		body        string
	}{
		{name: "root serves the index", path: "/", contentType: fiber.MIMETextHTML, body: "<html>app</html>"},
		{name: "asset is served as is", path: "/assets/app.js", contentType: fiber.MIMETextJavaScript, body: "console.log('app')"},
		{name: "client-side route falls back to the index", path: "/behandelaar/aanvraag/1", contentType: fiber.MIMETextHTML, body: "<html>app</html>"},
		{name: "unknown asset falls back to the index", path: "/assets/missing.js", contentType: fiber.MIMETextHTML, body: "<html>app</html>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, contentType, _, body := get(t, app, tt.path)

			assert.Equal(t, fiber.StatusOK, status)
			assert.Equal(t, tt.contentType, contentType)
			assert.Equal(t, tt.body, body)
		})
	}
}

func TestStaticAssetsRevalidate(t *testing.T) {
	t.Parallel()

	app := newMainApp(t, nil)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/assets/app.js", http.NoBody))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	tag := resp.Header.Get(fiber.HeaderETag)
	require.NotEmpty(t, tag)

	req := httptest.NewRequest(fiber.MethodGet, "/assets/app.js", http.NoBody)
	req.Header.Set(fiber.HeaderIfNoneMatch, tag)

	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Equal(t, fiber.StatusNotModified, resp.StatusCode)
}

func TestNewServiceRequiresIndex(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	assets := fstest.MapFS{".gitkeep": {Data: []byte{}}}

	srv, err := NewService(&config.Config{}, logger, assets, nil)

	require.ErrorContains(t, err, "the static assets have no index.html")
	assert.Nil(t, srv)
}
