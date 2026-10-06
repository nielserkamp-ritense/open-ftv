package network

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// transport is built once per request, so scheduled calls reuse its connections.
// A failed build is not kept, so a later call can still succeed, e.g. once a certificate is mounted.
func (r *Request) transport() (http.RoundTripper, error) {
	r.transportMutex.Lock()
	defer r.transportMutex.Unlock()

	if r.roundTripper == nil {
		rt, err := r.newTransport()
		if err != nil {
			return nil, err
		}

		r.roundTripper = rt
	}

	return r.roundTripper, nil
}

func (r *Request) newTransport() (http.RoundTripper, error) {
	if !r.usesTLS() {
		return http.DefaultTransport, nil
	}

	if err := r.validateTLS(); err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: r.Insecure, //nolint:gosec // the request config asks for it explicitly.
	}

	if r.CAFile != "" {
		pool, err := loadCA(r.CAFile)
		if err != nil {
			return nil, err
		}

		cfg.RootCAs = pool
	}

	if r.CertFile != "" {
		certFile := r.CertFile
		keyFile := r.KeyFile

		if _, err := loadKeyPair(certFile, keyFile); err != nil {
			return nil, err
		}

		// read the files again on every handshake, so a rotated certificate is picked up without a restart.
		cfg.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return loadKeyPair(certFile, keyFile)
		}
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = cfg

	return t, nil
}

func (r *Request) usesTLS() bool {
	return r.CAFile != "" || r.CertFile != "" || r.KeyFile != "" || r.Insecure
}

func (r *Request) validateTLS() error {
	if r.CAFile != "" && r.Insecure {
		return errors.New("tlsCA cannot be combined with insecure")
	}

	if (r.CertFile == "") != (r.KeyFile == "") {
		return errors.New("tlsCert and tlsKey must be set together")
	}

	return nil
}

func loadCA(path string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tlsCA: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no certificates found in tlsCA %s", path)
	}

	return pool, nil
}

func loadKeyPair(certFile, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load tlsCert and tlsKey: %w", err)
	}

	return &cert, nil
}
