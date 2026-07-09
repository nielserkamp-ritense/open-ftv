package server

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

// StatisticsResponse is the JSON body returned by the statistics endpoint and
// carried as the data of a statistics CloudEvent. It contains no personal data:
// only the aggregation parameters and per-(afnemer,doel,periode) tallies.
type StatisticsResponse struct {
	Vanaf        *time.Time    `json:"vanaf,omitempty"`
	Tot          *time.Time    `json:"tot,omitempty"`
	Bucket       query.Bucket  `json:"bucket"`
	K            int           `json:"k"`
	Statistieken []query.Count `json:"statistieken"`
}

// handleStatistieken serves GET /v1/statistieken.
//
// Query parameters: vanaf, tot (RFC3339 or YYYY-MM-DD), doel, bucket
// (day/week/month), event (event_name filter; defaults to adl.access_evaluation
// so statistics cover data-access decisions, not Inzicht-API meta-access).
func (s *Service) handleStatistieken(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from, err := parseTime(q.Get("vanaf"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'vanaf': "+err.Error())
		return
	}
	to, err := parseTime(q.Get("tot"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'tot': "+err.Error())
		return
	}

	event := q.Get("event")
	if event == "" {
		event = adl.EventAccessEvaluation
	}

	resp, err := s.aggregate(r.Context(), from, to, q.Get("doel"), event, s.bucket(q.Get("bucket")))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Access to the aggregate statistics is itself logged in the ADL, under the
	// authenticated caller identity (never the spoofable X-Verstrekker header).
	s.logAccess(r.Context(), "statistieken", callerVerstrekker(r), true, map[string]any{
		"doel":   q.Get("doel"),
		"bucket": string(resp.Bucket),
	})

	writeJSON(w, http.StatusOK, resp)
}

// aggregate runs the shared aggregation used by both the endpoint and the push.
func (s *Service) aggregate(ctx context.Context, from, to *time.Time, doel, event string, bucket query.Bucket) (StatisticsResponse, error) {
	f := query.Filter{Doel: doel, EventName: event}
	if from != nil {
		f.From = *from
	}
	if to != nil {
		f.To = *to
	}

	records, err := s.source.Query(ctx, f)
	if err != nil {
		return StatisticsResponse{}, err
	}

	counts := query.Aggregate(records, query.AggregateOptions{Bucket: bucket, K: s.kThreshold()})
	return StatisticsResponse{
		Vanaf:        from,
		Tot:          to,
		Bucket:       bucket,
		K:            s.kThreshold(),
		Statistieken: counts,
	}, nil
}

// parseTime accepts RFC3339 or a plain date (YYYY-MM-DD); "" yields nil.
func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		u := t.UTC()
		return &u, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	u := t.UTC()
	return &u, nil
}

// callerVerstrekker returns the verstrekker identity of the authenticated caller
// (from the bearer token, via the request context), used for access logging. It
// deliberately replaces the old X-Verstrekker header lookup, which was spoofable.
func callerVerstrekker(r *http.Request) string {
	if p, ok := principalFrom(r.Context()); ok {
		return p.Verstrekker
	}
	return ""
}
