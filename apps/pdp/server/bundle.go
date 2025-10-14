package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (s *Services) getLatestBundle(url string) (*http.Response, error) {
	b := &bundleGetter{url: url, ctx: s.ctx, cfg: s.cfg}

	if err := b.prepareTransport(); err != nil {
		return nil, err
	}

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

func (b *bundleGetter) prepareTransport() error {
	// b.transport = http.DefaultTransport.(*http.Transport).Clone()

	cfg := &tls.Config{}

	if b.cfg.BundleCA != "" {
		caCert, err := os.ReadFile(b.cfg.BundleCA)
		if err != nil {
			return err
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("invalid CA certificate")
		}

		cfg.RootCAs = certPool
	}

	if b.cfg.BundleCert != "" && b.cfg.BundleKey != "" {
		cert, err := tls.LoadX509KeyPair(b.cfg.BundleCert, b.cfg.BundleKey)
		if err != nil {
			return err
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	b.transport = &http.Transport{TLSClientConfig: cfg}
	b.transport.DisableCompression = true

	return nil
}

func (b *bundleGetter) prepareClient() {
	b.client = &http.Client{
		Transport: b.transport,
		Timeout:   b.cfg.BundleTimeout,
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
		b.req.Header.Add(models.HeaderAPIKey, b.cfg.BundleAPIKey)
	}

	if b.cfg.BundleEncoding != "" {
		b.req.Header.Add(fiber.HeaderContentEncoding, b.cfg.BundleEncoding)
	}
	return
}

func (b *bundleGetter) doRequest() {
	// ctx, cancel := context.WithTimeout(b.ctx, b.cfg.BundleTimeout)
	// defer cancel()

	b.resp, b.err = b.client.Do(b.req)
	if b.err != nil {
		return
	}

	if b.resp.StatusCode != http.StatusOK {
		b.resp.Body.Close()
		b.err = fmt.Errorf("failed to retrieve bundle: status %d", b.resp.StatusCode)
	}
}
