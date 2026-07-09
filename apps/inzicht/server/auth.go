package server

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
)

// Role is the authorization role carried by an Inzicht bearer token.
type Role string

// The Inzicht roles.
const (
	// RoleVerstrekker may create verzoeken and retrieve the result of its own
	// (approved) verzoeken, and read aggregate statistics.
	RoleVerstrekker Role = "verstrekker"
	// RoleBeheerder is the afnemer-side approver: it lists pending verzoeken and
	// approves/denies them, and reads aggregate statistics.
	RoleBeheerder Role = "beheerder"
)

// principal is the authenticated caller resolved from the bearer token (or, when
// auth is disabled for local play, from the X-Verstrekker header). It is the
// authoritative identity: it replaces the previously spoofable X-Verstrekker header
// for every authorization and access-logging decision.
type principal struct {
	Role        Role
	Verstrekker string
}

type ctxKey int

const principalKey ctxKey = iota

// principalFrom returns the authenticated principal stored in the request context.
func principalFrom(ctx context.Context) (*principal, bool) {
	p, ok := ctx.Value(principalKey).(*principal)
	return p, ok
}

// parseAuthTokens parses the "token:role:verstrekker" comma-separated token list
// into a token->principal map. Malformed or role-less entries are skipped.
func parseAuthTokens(spec string) map[string]*principal {
	out := map[string]*principal{}
	for _, entry := range strings.Split(spec, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) < 2 {
			continue
		}
		token := strings.TrimSpace(parts[0])
		role := Role(strings.TrimSpace(parts[1]))
		if token == "" || (role != RoleVerstrekker && role != RoleBeheerder) {
			continue
		}
		var verstrekker string
		if len(parts) == 3 {
			verstrekker = strings.TrimSpace(parts[2])
		}
		out[token] = &principal{Role: role, Verstrekker: verstrekker}
	}
	return out
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}

// authenticate is the HTTP middleware that resolves the caller identity from the
// bearer token (or, in AuthDisabled mode, the X-Verstrekker header) and stores it
// on the request context. Unauthenticated requests to protected routes get 401.
// The health endpoint is always open.
func (s *Service) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		p, ok := s.resolvePrincipal(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "unauthorized: valid bearer token required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey, p)))
	})
}

// resolvePrincipal maps an incoming request to an authenticated principal.
//
//   - AuthDisabled: local-play fallback - identity from X-Verstrekker (spoofable),
//     with full (beheerder) capability so every flow can be exercised locally.
//   - otherwise: the bearer token must match a configured token. With no tokens
//     configured the map is empty and every request is refused (fail-closed 401).
func (s *Service) resolvePrincipal(r *http.Request) (*principal, bool) {
	if s.authDisabled {
		v := verstrekkerHeader(r)
		if v == "" {
			v = "local"
		}
		return &principal{Role: RoleBeheerder, Verstrekker: v}, true
	}

	tok := bearerToken(r)
	if tok == "" {
		return nil, false
	}
	for t, p := range s.tokens {
		if subtle.ConstantTimeCompare([]byte(t), []byte(tok)) == 1 {
			return p, true
		}
	}
	return nil, false
}

// requireBeheerder writes 403 and returns false when the caller is not a beheerder.
func requireBeheerder(w http.ResponseWriter, p *principal) bool {
	if p == nil || p.Role != RoleBeheerder {
		writeError(w, http.StatusForbidden, "forbidden: beheerder role required")
		return false
	}
	return true
}

// verstrekkerHeader reads the legacy X-Verstrekker header (used only as the
// local-play identity fallback when auth is disabled).
func verstrekkerHeader(r *http.Request) string {
	if v := r.Header.Get("X-Verstrekker"); v != "" {
		return v
	}
	return r.URL.Query().Get("verstrekker")
}
