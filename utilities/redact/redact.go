// Package redact strips network-identifying details from free-text error messages.
package redact

import "regexp"

const mask = "*****"

var (
	// bracketedIPv6Regex matches the bracketed form Go's net package uses for IPv6 hosts in dial
	// errors, e.g. "[::1]" or "[2001:db8::1]:8080".
	bracketedIPv6Regex = regexp.MustCompile(`\[[0-9a-fA-F:]*:[0-9a-fA-F:]*](:\d+)?`)
	// bareIPv6Regex is a defense-in-depth fallback for unbracketed IPv6 addresses.
	bareIPv6Regex = regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{1,4}\b`)
	// ipv4Regex matches dotted-quad addresses, with an optional trailing port.
	ipv4Regex = regexp.MustCompile(`\b\d{1,3}(\.\d{1,3}){3}(:\d+)?\b`)
	// urlRegex matches absolute URLs of any scheme, stopping at whitespace or a closing quote
	// (Go's *url.Error wraps the URL in double quotes, e.g. `Get "http://...": ...`).
	urlRegex = regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9+.-]*://[^\s"]+`)
	// hostRegex matches bare hostnames (no scheme), e.g. as seen in DNS lookup errors.
	hostRegex = regexp.MustCompile(`\b([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}\b`)
)

// Sensitive redacts IPv4/IPv6 addresses, URLs, and hostnames from input, e.g. so a dial error
// against a user-supplied URL can't be used to fingerprint internal network topology.
func Sensitive(input string) string {
	out := bracketedIPv6Regex.ReplaceAllString(input, mask)
	out = bareIPv6Regex.ReplaceAllString(out, mask)
	out = ipv4Regex.ReplaceAllString(out, mask)
	out = urlRegex.ReplaceAllString(out, mask)
	out = hostRegex.ReplaceAllString(out, mask)

	return out
}
