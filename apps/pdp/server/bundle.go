package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

func (s *service) getLatestBundle(url string) (*http.Response, error) {
	b := &bundleGetter{url: url, ctx: s.ctx, cfg: s.cfg}
	b.prepareClient()
	b.prepareRequest()

	if b.err != nil {
		return nil, b.err
	}

	retries := s.cfg.BundleRetries
	count := retries
	backoff := time.Second * 5

	for {
		if b.doRequest(); b.err == nil {
			return b.resp, nil
		}

		count--

		if count > 0 {
			s.logger.Warn("failed to retrieve bundle", "error", b.err, "retry in", backoff.String())
		} else {
			s.logger.Warn("failed to retrieve bundle", "error", b.err)
		}

		if count <= 0 {
			return nil, fmt.Errorf("failed to retrieve bundle; tried %d times", retries)
		}

		time.Sleep(backoff)
		backoff *= 2
	}
}

type bundleGetter struct {
	url       string
	ctx       context.Context
	cfg       *config.Config
	transport *http.Transport
	client    *http.Client
	req       *http.Request
	resp      *http.Response
	err       error
}

func (b *bundleGetter) prepareClient() {
	b.transport = http.DefaultTransport.(*http.Transport).Clone()
	b.transport.DisableCompression = true

	b.client = &http.Client{
		Transport:     b.transport,
		CheckRedirect: http.DefaultClient.CheckRedirect,
		Jar:           http.DefaultClient.Jar,
		Timeout:       b.cfg.BundleTimeout,
	}
}

func (b *bundleGetter) prepareRequest() {
	b.req, b.err = http.NewRequest("GET", b.url, nil)
	if b.err != nil {
		return
	}

	if b.cfg.BundleHeaders != "" {
		list := strings.Split(b.cfg.BundleHeaders, ",")
		for i := range list {
			parts := strings.Split(list[i], ":")
			if len(parts) == 2 {
				b.req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	if b.cfg.BundleAPIKey != "" {
		b.req.Header.Add("ApiKey", b.cfg.BundleAPIKey)
	}

	if b.cfg.BundleEncoding != "" {
		b.req.Header.Add(fiber.HeaderContentEncoding, b.cfg.BundleEncoding)
	}
	return
}

func (b *bundleGetter) doRequest() {
	ctx, cancel := context.WithTimeout(b.ctx, b.cfg.BundleTimeout)
	defer cancel()

	b.resp, b.err = b.client.Do(b.req.WithContext(ctx))
	if b.err != nil {
		return
	}

	if b.resp.StatusCode != http.StatusOK {
		b.resp.Body.Close()
		b.err = fmt.Errorf("failed to retrieve bundle: status %d", b.resp.StatusCode)
	}
}
