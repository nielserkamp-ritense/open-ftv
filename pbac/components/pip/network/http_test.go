package network

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/server"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestRunner_InitClient(t *testing.T) {
	g := &getAttr{m: make(map[string]models.Attribute)}

	_, caFile, caCert, caKey := makeIntermediate(t)
	certFile, keyFile := makeCert(t, caCert, caKey)

	testCases := []struct {
		name       string
		req        *Request
		want       bool
		wantClient bool
	}{
		{
			name: "bad CertFile & KeyFile",
			req: &Request{
				Name:     "req1",
				Timeout:  time.Second,
				CertFile: "/not/a/valid/file",
				KeyFile:  "/not/a/valid/file",
			},
		},
		{
			name: "no TLS",
			req: &Request{
				Name:    "req1",
				Timeout: time.Second,
			},
			want:       true,
			wantClient: true,
		},
		{
			name: "with TLS",
			req: &Request{
				Name:     "req1",
				Timeout:  time.Second,
				CAFile:   caFile,
				CertFile: certFile,
				KeyFile:  keyFile,
			},
			want:       true,
			wantClient: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			tc.req.prepare()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			m := &manager{ctx: ctx, cancel: cancel, logger: logger, get: g}
			r := &runner{manager: m, ctx: ctx, logger: logger, req: tc.req}

			got := r.initClient()
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantClient, r.httpClient != nil)
		})
	}
}

func TestManager_Execute(t *testing.T) {
	var data = []byte(`[
 {"id":"dfa58ae2-0dd4-471d-8863-08147944e641","oin":"01726477373371538205","attributes":{"name":"FSC controller","isMember":true,"maturity":4}},
 {"id":"f8d73ae9-f3a7-4521-a6c0-28422b056cbb","oin":"01726469521943092351","attributes":{"name":"RDW","isMember":true,"maturity":3}},
 {"id":"86af57a8-c9d6-4620-8511-e81498ff2df7","oin":"01726469365987449994","attributes":{"name":"RViG","isMember":true,"maturity":5}}
]`)

	handler := func(req *fiber.Ctx) error {
		req.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return req.Send(data)
	}

	g := &getAttr{m: make(map[string]models.Attribute)}

	_, caFile, caCert, caKey := makeIntermediate(t)
	certFile1, keyFile1 := makeCert(t, caCert, caKey)
	certFile2, keyFile2 := makeCert(t, caCert, caKey)

	h1 := slog2.NewDummyHandler(slog.LevelDebug)
	l1 := slog.New(h1)
	s1 := newService(t, l1, "", "", "", "ledenlijst", handler)

	h2 := slog2.NewDummyHandler(slog.LevelDebug)
	l2 := slog.New(h2)
	s2 := newService(t, l2, "", certFile1, keyFile1, "ledenlijst", handler)

	h3 := slog2.NewDummyHandler(slog.LevelDebug)
	l3 := slog.New(h3)
	s3 := newService(t, l3, caFile, certFile1, keyFile1, "ledenlijst", handler)

	testCases := []struct {
		name   string
		logger *slog.Logger
		svc    server.Service
		req    *Request
	}{
		{
			name:   "FDS ledenlijst - no TLS",
			logger: l1,
			svc:    s1,
			req: &Request{
				Name:    "FDS ledenlijst",
				Method:  "GET",
				URI:     "http://127.0.0.1:9000/v1/ledenlijst",
				Timeout: time.Minute,
			},
		},
		{
			name:   "FDS ledenlijst - server TLS",
			logger: l2,
			svc:    s2,
			req: &Request{
				Name:    "FDS ledenlijst",
				Method:  "GET",
				URI:     "https://127.0.0.1:9000/v1/ledenlijst",
				CAFile:  caFile,
				Timeout: time.Minute,
			},
		},
		{
			name:   "FDS ledenlijst - mTLS",
			logger: l3,
			svc:    s3,
			req: &Request{
				Name:     "FDS ledenlijst",
				Method:   "GET",
				URI:      "https://127.0.0.1:9000/v1/ledenlijst",
				CAFile:   caFile,
				CertFile: certFile2,
				KeyFile:  keyFile2,
				Timeout:  time.Minute,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wg.Add(2)

			go func(wg *sync.WaitGroup) {
				tc.svc.Serve()
				wg.Done()
			}(&wg)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			m := &manager{ctx: ctx, cancel: cancel, logger: tc.logger, get: g}

			var execErr error
			go func(wg *sync.WaitGroup) {
				time.Sleep(25 * time.Millisecond)

				execErr = m.execute(tc.logger, tc.req)
				tc.svc.Shutdown()
				wg.Done()
			}(&wg)

			wg.Wait()

			require.NoError(t, execErr)

			h1.Clear()
			h2.Clear()
		})
	}
}

type getAttr struct {
	m map[string]models.Attribute
}

func (g *getAttr) GetAttribute(name string) models.Attribute {
	return g.m[name]
}
