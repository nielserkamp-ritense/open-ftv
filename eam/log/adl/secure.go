package adl

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// AllowInsecureTransport gates the ADL "Encryption" requirement: the standard mandates that
// the log MUST enforce TLS on its connections. When false (the default) a sink endpoint that
// would carry records in cleartext over the network - an http:// scheme - is rejected at
// construction with a clear error. An operator can opt out by setting this to true, which the
// application wires from the ADL_INSECURE config flag (eam/config.ADL.Insecure); it exists for
// development and closed-network deployments only. Set it once at startup, before the sinks
// are constructed.
//
// Loopback endpoints (localhost, 127.0.0.0/8, ::1) are always permitted regardless of this
// flag: traffic to loopback never leaves the host, so terminating TLS at a co-located
// collector/sidecar reached over loopback is a deployment choice, not a network-exposure
// risk. Non-URL targets (stdout, stderr, slog, or a bare host:port whose transport handles
// its own security) are likewise left to their own transport.
var AllowInsecureTransport = false

// requireSecureEndpoint returns an error when endpoint uses a cleartext http:// scheme on a
// non-loopback host and insecure transport has not been explicitly allowed. It is a no-op for
// https, loopback, and non-http targets (stdout/stderr/slog/host:port).
func requireSecureEndpoint(endpoint string) error {
	e := strings.TrimSpace(endpoint)
	if e == "" {
		return nil
	}
	if !strings.HasPrefix(strings.ToLower(e), "http://") {
		return nil // https, or a non-cleartext-URL target - not our concern.
	}
	if AllowInsecureTransport {
		return nil
	}
	if u, err := url.Parse(e); err == nil && isLoopbackHost(u.Hostname()) {
		return nil
	}
	return fmt.Errorf("adl: refusing insecure cleartext endpoint %q: the ADL standard requires "+
		"TLS on log connections; use https:// or set ADL_INSECURE=true to opt out", endpoint)
}

// isLoopbackHost reports whether host is localhost or a loopback IP.
func isLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
