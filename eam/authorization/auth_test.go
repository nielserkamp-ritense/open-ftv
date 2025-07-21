package authorization

import (
	"context"
	"log/slog"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
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
	ip := pip.New(ctx, log, pip.WithFileStore("../../testdata/unittest/auth", true), pip.WithFactories(models.NewAttributeSet, models.NewEntitySet))
	ap := pap.New(ctx, log, pap.WithLanguage("cedar"), pap.WithFileStore("../../testdata/unittest/auth/policies", true))
	dp := cedar_embedded.NewController(pdp.WithLogger(log), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithPEP(ep))

	authenticator := authentication.NewBCrypt(authentication.WithContext(ctx), authentication.WithLogger(log), authentication.WithEntityGetter(ip.GetEntity))
	require.NotNil(t, authenticator)

	parseURL := func(s string) *url.URL {
		u, _ := url.Parse(s)
		return u
	}

	testCases := []struct {
		name      string
		req       *Request
		wantErr   bool
		wantAllow bool
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
			wantAllow: false,
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
			wantAllow: true,
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
			wantAllow: false,
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
			wantAllow: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := New(WithContext(ctx), WithLogger(log), WithPEP(ep), WithEntityGetter(ip.GetEntity), WithPDP(dp), WithAuthenticator(authenticator))
			require.NotNil(t, a)

			uid := uuid.New()
			tc.req.UID = &uid

			got, err2 := a.Authorize(tc.req)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, got)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantAllow, got.Allowed)
			}
		})
	}
}
