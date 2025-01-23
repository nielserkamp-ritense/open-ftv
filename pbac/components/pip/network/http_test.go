package network

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestRunner_InitClient(t *testing.T) {
	g := &getAttr{m: make(map[string]models.Attribute)}

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
	g := &getAttr{m: make(map[string]models.Attribute)}

	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "FDS ledenlijst",
			req: &Request{
				Name:    "FDS ledenlijst",
				Method:  "GET",
				URI:     "http://localhost:8090/v1/leden",
				Timeout: time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			m := &manager{ctx: ctx, cancel: cancel, logger: logger, get: g}

			m.execute(logger, tc.req)
		})
	}
}

type getAttr struct {
	m map[string]models.Attribute
}

func (g *getAttr) GetAttribute(name string) models.Attribute {
	return g.m[name]
}
