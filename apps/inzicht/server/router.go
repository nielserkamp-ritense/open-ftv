package server

import (
	"encoding/json"
	"net/http"
)

// routes builds the Inzicht routing table using net/http method-aware patterns.
func (s *Service) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Aggregate usage statistics per doel/afnemer (shared with the verstrekker).
	mux.HandleFunc("GET /v1/statistieken", s.handleStatistieken)

	// Inzicht API with an approval step on the afnemer side.
	mux.HandleFunc("POST /v1/inzicht/verzoeken", s.handleCreateVerzoek)
	mux.HandleFunc("GET /v1/inzicht/verzoeken", s.handleListVerzoeken)
	mux.HandleFunc("POST /v1/inzicht/verzoeken/{id}/approve", s.handleApprove)
	mux.HandleFunc("POST /v1/inzicht/verzoeken/{id}/deny", s.handleDeny)
	mux.HandleFunc("GET /v1/inzicht/verzoeken/{id}/resultaat", s.handleResultaat)

	return mux
}

// writeJSON writes v as an indented JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// writeError writes a JSON error envelope.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
