package authorization

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	log := slog.New(h)

	p := pip.New(ctx, log, pip.WithFileStore("../../testdata/pip", false))

	p2 := pep.New(ctx, log)
	p3 := cedar_embedded.NewController(pdp.WithLogger(log))

	testCases := []struct {
		name    string
		opts    []Option
		wantCtx context.Context
		wantLog *slog.Logger
		wantPEP *pep.PEP
		wantPDP pdp.Controller
	}{
		{name: "no options"},
		{name: "context", opts: []Option{WithContext(ctx)}, wantCtx: ctx},
		{name: "logger", opts: []Option{WithLogger(log)}, wantLog: log},
		{name: "pep", opts: []Option{WithPEP(p2)}, wantPEP: p2},
		{name: "pdp", opts: []Option{WithPDP(p3)}, wantPDP: p3},
		{name: "entities", opts: []Option{WithEntityGetter(p.GetEntity)}},
		{name: "all", opts: []Option{WithPEP(p2), WithLogger(log), WithPDP(p3), WithEntityGetter(p.GetEntity), WithContext(ctx)}, wantCtx: ctx, wantLog: log, wantPEP: p2, wantPDP: p3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := New(tc.opts...)
			require.NotNil(t, got)

			got2, ok := got.(*auth)
			require.True(t, ok)
			require.NotNil(t, got2)

			assert.NotNil(t, got2.ctx)
			assert.NotNil(t, got2.log)
			assert.NotNil(t, got2.pep)
			assert.NotNil(t, got2.getter)

			if tc.wantCtx != nil {
				assert.Equal(t, tc.wantCtx, got2.ctx)
			}
			if tc.wantLog != nil {
				assert.Equal(t, tc.wantLog, got2.log)
			}
			if tc.wantPEP != nil {
				assert.Equal(t, tc.wantPEP, got2.pep)
			}
			if tc.wantPDP != nil {
				assert.Equal(t, tc.wantPDP, got2.pdp)
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	log := slog.New(h)

	ep := pep.New(ctx, log)
	ip := pip.New(ctx, log, pip.WithFileStore("../../testdata/unittest/auth", true))
	ap := pap.New(ctx, log, pap.WithLanguage("cedar"), pap.WithFileStore("../../testdata/unittest/auth/policies", true))
	dp := cedar_embedded.NewController(pdp.WithLogger(log), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithPEP(ep))

	authenticator := authentication.NewBCrypt(authentication.WithLogger(log), authentication.WithEntityGetter(ip.GetEntity))
	require.NotNil(t, authenticator)

	parseURL := func(s string) *url.URL {
		u, _ := url.Parse(s)
		return u
	}

	testCases := []struct {
		name          string
		req           *Request
		wantErr       bool
		wantAllow     bool
		wantPrincipal identity.Principal
	}{
		{
			name: "no authentication",
			req: &Request{
				URL:     parseURL("https://openftv.nl/v1/attributes"),
				Method:  "GET",
				Headers: map[string][]string{},
			},
			wantErr: true,
		},
		{
			name: "wrong user",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "GET",
				Headers: map[string][]string{
					// echo -n 'minnie:mouse' | base64
					"Authorization": {"Basic bWlubmllOm1vdXNl"},
				},
			},
			wantErr: true,
		},
		{
			name: "wrong password",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "GET",
				Headers: map[string][]string{
					// echo -n 'mickey:mousse' | base64
					"Authorization": {"Basic bWlja2V5Om1vdXNzZQ=="},
				},
			},
			wantErr: true,
		},
		{
			name: "not admin",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "POST",
				Headers: map[string][]string{
					// echo -n 'mickey:mouse' | base64
					"Authorization": {"Basic bWlja2V5Om1vdXNl"},
				},
			},
			wantAllow:     false,
			wantPrincipal: identity.NewUnknownPrincipal(),
		},
		{
			name: "is admin",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "POST",
				Headers: map[string][]string{
					// echo -n 'admin:admin' | base64
					"Authorization": {"Basic YWRtaW46YWRtaW4="},
				},
			},
			wantAllow:     true,
			wantPrincipal: identity.NewPrincipal(identity.KindUser, "admin"),
		},
		{
			name: "wrong apikey",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "GET",
				Headers: map[string][]string{
					"Api-Key": {"123456"},
				},
			},
			wantErr: true,
		},
		{
			name: "good apikey - bad method",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "POST",
				Headers: map[string][]string{
					"Api-Key": {"abcdef"},
				},
			},
			wantAllow:     false,
			wantPrincipal: identity.NewUnknownPrincipal(),
		},
		{
			name: "good apikey - good method",
			req: &Request{
				URL:    parseURL("https://openftv.nl/v1/attributes"),
				Method: "GET",
				Headers: map[string][]string{
					"Api-Key": {"abcdef"},
				},
			},
			// API-key callers are deliberately NOT attributed as a user (scoped out
			// to avoid leaking the raw key into audit-read endpoints) — they fall
			// back to the system principal like any other non-user caller.
			wantPrincipal: identity.NewSystemPrincipal(),
			wantAllow:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := New(WithContext(ctx), WithLogger(log), WithPEP(ep), WithEntityGetter(ip.GetEntity), WithPDP(dp), WithAuthenticator(authenticator))
			require.NotNil(t, a)

			uid := uuid.New()
			tc.req.UID = &uid

			got, gotPrincipal, err2 := a.Authorize(tc.req)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, got)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantAllow, got.Allowed)

				if tc.wantPrincipal != identity.NewUnknownPrincipal() {
					assert.Equal(t, tc.wantPrincipal, gotPrincipal)
				}
			}
		})
	}
}

// A caller who only presents a JWT (no Basic-Auth or API key) should still count as a user.
func TestAuthorize_JWTPrincipal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	log := slog.New(h)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	kf := func(*jwt.Token) (any, error) { return &key.PublicKey, nil }

	ep := pep.New(ctx, log, pep.WithJWT(pep.JWTConfig{
		Keyfunc:  kf,
		Issuer:   "https://idp.test",
		Audience: "openftv",
	}))

	a := New(WithContext(ctx), WithLogger(log), WithPEP(ep), NoAuth())
	require.NotNil(t, a)

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://idp.test",
		"aud": "openftv",
		"sub": "alice@wonderland.cc",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	bearer, err2 := tok.SignedString(key)
	require.NoError(t, err2)

	u, _ := url.Parse("https://openftv.nl/v1/policy/123")
	uid := uuid.New()

	resp, principal, err3 := a.Authorize(&Request{
		UID:    &uid,
		URL:    u,
		Method: "PUT",
		Headers: map[string][]string{
			"Authorization": {"Bearer " + bearer},
		},
	})

	require.NoError(t, err3)
	require.NotNil(t, resp)
	assert.Equal(t, identity.KindUser, principal.Kind)
	assert.Equal(t, "alice@wonderland.cc", principal.ID)
}
