package fiber

import (
	"fmt"
	"syscall"
	"time"
)

func (s *service) run() {
	t := "http service"
	if s.MutualTLS || (s.TLSCert != "" && s.TLSKey != "") {
		t = "https service"
	}
	s.logger.Info(fmt.Sprintf("%s starting", t), "host", s.Host, "port", s.Port)

	go func() {
		// there's no notification from fiber when the HTTP server has started, so we fake it.
		// 10 milliseconds should be ample time.
		time.AfterFunc(10*time.Millisecond, func() {
			s.logger.Info(fmt.Sprintf("%s started", t))
		})

		address := fmt.Sprintf("%s:%d", s.Host, s.Port)

		var err error

		switch {
		case s.MutualTLS:
			err = s.svc.ListenMutualTLS(address, s.TLSCert, s.TLSKey, s.CA)
		case s.TLSCert != "" && s.TLSKey != "":
			err = s.svc.ListenTLS(address, s.TLSCert, s.TLSKey)
		default:
			err = s.svc.Listen(address)
		}

		if err != nil {
			s.logger.Error(fmt.Sprintf("failed to start %s", t), "error", err)
			s.shutdown.Store(true)       // indicate server is stopped or never started.
			s.intChan <- syscall.SIGQUIT // send a signal to end the Run function.
		}
	}()

	// first signal performs orderly shutdown.
	<-s.intChan

	if s.shutdown.Load() {
		// this indicates the server never got started, so there's nothing to shut down!
		return
	}

	go func() {
		s.logger.Info(fmt.Sprintf("%s stopping", t))
		s.cancel()
		_ = s.svc.Shutdown()
		s.shutdown.Store(true)       // indicate shutdown was successful.
		s.intChan <- syscall.SIGQUIT // send a second signal to end the Run function.
	}()

	// another signal will force the shutdown.
	<-s.intChan

	if s.shutdown.Load() {
		s.logger.Info(fmt.Sprintf("%s stopped successfully", t))
	} else {
		s.logger.Info(fmt.Sprintf("%s aborted", t))
	}
}
