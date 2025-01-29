package network

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/server"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestRunner_InitClient(t *testing.T) {
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

			m := &manager{ctx: ctx, cancel: cancel, logger: logger}
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

	handler1 := func(req *fiber.Ctx) error {
		req.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return req.Send(data)
	}

	handler2 := func(req *fiber.Ctx) error {
		return req.SendStatus(fiber.StatusNotModified)
	}

	handler3 := func(req *fiber.Ctx) error {
		return req.SendStatus(fiber.StatusInternalServerError)
	}

	_, caFile, caCert, caKey := makeIntermediate(t)
	certFile1, keyFile1 := makeCert(t, caCert, caKey)
	certFile2, keyFile2 := makeCert(t, caCert, caKey)

	h1 := slog2.NewDummyHandler(slog.LevelDebug)
	l1 := slog.New(h1)
	s1 := newService(t, l1, "", "", "", "ledenlijst", handler1)

	h2 := slog2.NewDummyHandler(slog.LevelDebug)
	l2 := slog.New(h2)
	s2 := newService(t, l2, "", certFile1, keyFile1, "ledenlijst", handler1)

	h3 := slog2.NewDummyHandler(slog.LevelDebug)
	l3 := slog.New(h3)
	s3 := newService(t, l3, caFile, certFile1, keyFile1, "ledenlijst", handler1)

	h4 := slog2.NewDummyHandler(slog.LevelDebug)
	l4 := slog.New(h4)
	s4 := newService(t, l4, caFile, certFile1, keyFile1, "ledenlijst", handler2)

	h5 := slog2.NewDummyHandler(slog.LevelDebug)
	l5 := slog.New(h5)
	s5 := newService(t, l5, caFile, certFile1, keyFile1, "ledenlijst", handler3)

	h6 := slog2.NewDummyHandler(slog.LevelDebug)
	l6 := slog.New(h6)
	s6 := newService(t, l6, caFile, certFile1, keyFile1, "ledenlijst", handler1)

	h7 := slog2.NewDummyHandler(slog.LevelDebug)
	l7 := slog.New(h7)
	s7 := newService(t, l7, caFile, certFile1, keyFile1, "ledenlijst", handler1)

	h8 := slog2.NewDummyHandler(slog.LevelDebug)
	l8 := slog.New(h8)
	s8 := newService(t, l8, caFile, certFile1, keyFile1, "ledenlijst", handler1)

	d1 := &ResponseMapping{}

	testCases := []struct {
		name    string
		logger  *slog.Logger
		svc     server.Service
		req     *Request
		wantErr bool
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
				Mapping: d1,
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
				Mapping: d1,
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
				Mapping:  d1,
			},
		},
		{
			name:   "FDS ledenlijst - mTLS - not modified",
			logger: l4,
			svc:    s4,
			req: &Request{
				Name:     "FDS ledenlijst",
				Method:   "GET",
				URI:      "https://127.0.0.1:9000/v1/ledenlijst",
				CAFile:   caFile,
				CertFile: certFile2,
				KeyFile:  keyFile2,
				Timeout:  time.Minute,
				Mapping:  d1,
			},
		},
		{
			name:   "FDS ledenlijst - mTLS - server error",
			logger: l5,
			svc:    s5,
			req: &Request{
				Name:     "FDS ledenlijst",
				Method:   "GET",
				URI:      "https://127.0.0.1:9000/v1/ledenlijst",
				CAFile:   caFile,
				CertFile: certFile2,
				KeyFile:  keyFile2,
				Timeout:  time.Minute,
				Mapping:  d1,
			},
			wantErr: true,
		},
		{
			name:   "bad URL",
			logger: l6,
			svc:    s6,
			req: &Request{
				Name:     "FDS ledenlijst",
				Method:   "GET",
				URI:      "https://127.0.0.1:9001/lijst",
				CAFile:   caFile,
				CertFile: certFile2,
				KeyFile:  keyFile2,
				Timeout:  200 * time.Millisecond,
				Mapping:  d1,
			},
			wantErr: true,
		},
		{
			name:   "bad ca",
			logger: l7,
			svc:    s7,
			req: &Request{
				Name:     "FDS ledenlijst",
				Method:   "GET",
				URI:      "https://127.0.0.1:9000/v1/ledenlijst",
				CertFile: certFile2,
				KeyFile:  keyFile2,
				Timeout:  100 * time.Millisecond,
				Mapping:  d1,
			},
			wantErr: true,
		},
		{
			name:   "bad method",
			logger: l8,
			svc:    s8,
			req: &Request{
				Name:     "FDS ledenlijst",
				Method:   "/001/001/001",
				URI:      "https://127.0.0.1:9000/v1/ledenlijst",
				CAFile:   caFile,
				CertFile: certFile2,
				KeyFile:  keyFile2,
				Timeout:  100 * time.Millisecond,
				Mapping:  d1,
			},
			wantErr: true,
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

			m := &manager{ctx: ctx, cancel: cancel, logger: tc.logger}

			var execErr error
			go func(wg *sync.WaitGroup) {
				time.Sleep(25 * time.Millisecond)

				execErr = m.execute(tc.logger, tc.req)
				tc.svc.Shutdown()
				wg.Done()
			}(&wg)

			wg.Wait()

			if tc.wantErr {
				require.Error(t, execErr)
			} else {
				require.NoError(t, execErr)
			}

			h1.Clear()
			h2.Clear()
		})
	}
}

func TestManager_Execute_WithDecode(t *testing.T) {
	var data = []byte(`[
 {"id":"dfa58ae2-0dd4-471d-8863-08147944e641","oin":"01726477373371538205","attributes":{"name":"FSC controller","isMember":true,"maturity":4}},
 {"id":"f8d73ae9-f3a7-4521-a6c0-28422b056cbb","oin":"01726469521943092351","attributes":{"name":"RDW","isMember":true,"maturity":3}},
 {"id":"86af57a8-c9d6-4620-8511-e81498ff2df7","oin":"01726469365987449994","attributes":{"name":"RViG","isMember":true,"maturity":5}}
]`)

	var decData = `
entities:
  - typeValue: fds_member
    idField: id
    attributes:
      - base: ""
        map:
          - keyValue: oin
            valueField: oin
            typeValue: xsd:string
      - base: attributes
        map:
          - keyValue: name
            valueField: name
            typeValue: xsd:string
          - keyValue: isMember
            valueField: isMember
            typeValue: xsd:boolean
          - keyValue: maturity
            valueField: maturity
            typeValue: xsd:short
`

	t.Run("execute with decoding", func(t *testing.T) {
		_, caFile, caCert, caKey := makeIntermediate(t)
		certFile1, keyFile1 := makeCert(t, caCert, caKey)
		certFile2, keyFile2 := makeCert(t, caCert, caKey)

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		handler := func(req *fiber.Ctx) error {
			req.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return req.Send(data)
		}

		svc := newService(t, logger, caFile, certFile1, keyFile1, "ledenlijst", handler)

		var dec ResponseMapping
		err := yaml.Unmarshal([]byte(decData), &dec)
		require.NoError(t, err)

		req := &Request{
			Name:        "ledenlijst",
			Description: "retrieve FDS ledenlijst",
			Method:      "GET",
			URI:         "https://127.0.0.1:9000/v1/ledenlijst",
			Headers:     map[string]string{fiber.HeaderAcceptEncoding: fiber.MIMEApplicationJSON},
			Timeout:     5 * time.Second,
			CAFile:      caFile,
			CertFile:    certFile2,
			KeyFile:     keyFile2,
			Mapping:     &dec,
		}

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func(wg *sync.WaitGroup) {
			svc.Serve()
			wg.Done()
		}(&wg)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		m := &manager{
			ctx:           ctx,
			cancel:        cancel,
			logger:        logger,
			entities:      models.NewEntitySet(),
			newAttributes: models.NewAttributeSet,
		}

		var execErr error
		go func(wg *sync.WaitGroup) {
			time.Sleep(25 * time.Millisecond)

			execErr = m.execute(logger, req)
			svc.Shutdown()
			wg.Done()
		}(&wg)

		wg.Wait()

		require.NoError(t, execErr)

		var count int
		m.entities.IterateEntities(func(e models.Entity) {
			count++

			assert.Equal(t, "fds_member", e.Type())

			switch e.ID() {
			case "dfa58ae2-0dd4-471d-8863-08147944e641":
				a := e.Attributes().GetAttribute("oin")
				require.NotNil(t, a)
				assert.Equal(t, "01726477373371538205", a.Value())

			case "f8d73ae9-f3a7-4521-a6c0-28422b056cbb":
				a := e.Attributes().GetAttribute("isMember")
				require.NotNil(t, a)
				assert.Equal(t, true, a.Value())

			case "86af57a8-c9d6-4620-8511-e81498ff2df7":
				a := e.Attributes().GetAttribute("maturity")
				require.NotNil(t, a)
				assert.Equal(t, int64(5), a.Value())
			}
		})

		assert.Equal(t, 3, count)
	})
}
