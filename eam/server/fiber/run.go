package fiber

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
)

func (s *service) run() {
	t := s.name()

	s.logger.Info(fmt.Sprintf("%s starting", t), "host", s.Host, "port", s.Port)

	// bind the port before waiting for a stop: a Shutdown that arrives earlier would otherwise find
	// nothing to stop, and the server would start listening after all.
	ln, err := s.listen()
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to start %s", t), "error", err)
		return
	}

	s.logger.Info(fmt.Sprintf("%s started", t))

	served := make(chan error, 1)

	go func() {
		served <- s.svc.Listener(ln)
	}()

	// first signal performs orderly shutdown.
	select {
	case <-s.intChan:
	case <-s.stop:
	case err := <-served:
		s.logger.Error(fmt.Sprintf("%s stopped serving", t), "error", err)
		return
	}

	s.logger.Info(fmt.Sprintf("%s stopping", t))

	stopped := make(chan struct{})

	go func() {
		s.cancel()
		_ = s.svc.Shutdown()
		// Shutdown finds nothing to stop while fiber is still taking over the listener; closing it ends Listener anyway.
		_ = ln.Close()

		close(stopped)
	}()

	// another signal will force the shutdown.
	select {
	case <-stopped:
		s.logger.Info(fmt.Sprintf("%s stopped successfully", t))
	case <-s.intChan:
		s.logger.Info(fmt.Sprintf("%s aborted", t))
	}
}

func (s *service) name() string {
	t := "http service"
	if s.MutualTLS || (s.TLSCert != "" && s.TLSKey != "") {
		t = "https service"
	}

	if s.SvcName != "" {
		t = fmt.Sprintf("%s (%s)", t, s.SvcName)
	}

	return t
}

// the TLS settings follow fiber's ListenTLS and ListenMutualTLS.
func (s *service) listen() (net.Listener, error) {
	address := fmt.Sprintf("%s:%d", s.Host, s.Port)

	if !s.MutualTLS && (s.TLSCert == "" || s.TLSKey == "") {
		return net.Listen(s.svc.Config().Network, address)
	}

	cert, err := tls.LoadX509KeyPair(s.TLSCert, s.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("tls: cannot load TLS key pair from certFile=%q and keyFile=%q: %w", s.TLSCert, s.TLSKey, err)
	}

	config := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}

	if s.MutualTLS {
		ca, err := os.ReadFile(s.CA)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = x509.NewCertPool()
		config.ClientCAs.AppendCertsFromPEM(ca)
	}

	return tls.Listen(s.svc.Config().Network, address, config)
}
