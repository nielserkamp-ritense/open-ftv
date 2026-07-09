package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/verzoek"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

// createRequest is the JSON body for POST /v1/inzicht/verzoeken.
//
// The verstrekker is NOT taken from this body (nor from a header): it is bound to
// the authenticated caller identity. An afnemer scope, however, is part of the
// requested (and to-be-approved) disclosure and is carried here.
type createRequest struct {
	Vanaf    *time.Time `json:"vanaf,omitempty"`
	Tot      *time.Time `json:"tot,omitempty"`
	Doel     string     `json:"doel,omitempty"`
	Afnemer  string     `json:"afnemer,omitempty"`
	TraceIDs []string   `json:"trace_ids,omitempty"`
}

// decideRequest is the JSON body for approve/deny (optional).
type decideRequest struct {
	DecidedBy string `json:"decided_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// handleCreateVerzoek serves POST /v1/inzicht/verzoeken: the verstrekker requests
// a set of processing activities. The request is stored with status pending.
func (s *Service) handleCreateVerzoek(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r.Context())
	// The verstrekker is bound to the authenticated identity, not to any client
	// input. Only the verstrekker role may create (own) a verzoek; a beheerder is an
	// approver, not a requester. The local-play opt-out bypasses the role check.
	if p == nil || p.Verstrekker == "" || (p.Role != RoleVerstrekker && !s.authDisabled) {
		writeError(w, http.StatusForbidden, "forbidden: the verstrekker role is required to create a verzoek")
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	v := verzoek.Verzoek{
		ID:          randomHex(8),
		Verstrekker: p.Verstrekker,
		Vanaf:       req.Vanaf,
		Tot:         req.Tot,
		Doel:        req.Doel,
		Afnemer:     req.Afnemer,
		TraceIDs:    req.TraceIDs,
		Status:      verzoek.StatusPending,
		CreatedAt:   now().UTC(),
	}

	created, err := s.store.Create(v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.logAccess(r.Context(), "verzoek.create", p.Verstrekker, true, map[string]any{
		"verzoek_id": created.ID,
		"doel":       created.Doel,
		"afnemer":    created.Afnemer,
		"trace_ids":  created.TraceIDs,
	})

	writeJSON(w, http.StatusCreated, created)
}

// handleListVerzoeken serves GET /v1/inzicht/verzoeken for the afnemer-beheerder:
// the approval queue. Only the beheerder role may review it.
func (s *Service) handleListVerzoeken(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r.Context())
	if !requireBeheerder(w, p) {
		return
	}
	list, err := s.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.logAccess(r.Context(), "verzoek.list", p.Verstrekker, true, nil)
	writeJSON(w, http.StatusOK, map[string]any{"verzoeken": list})
}

// handleApprove serves POST /v1/inzicht/verzoeken/{id}/approve.
func (s *Service) handleApprove(w http.ResponseWriter, r *http.Request) {
	s.decide(w, r, verzoek.StatusApproved)
}

// handleDeny serves POST /v1/inzicht/verzoeken/{id}/deny.
func (s *Service) handleDeny(w http.ResponseWriter, r *http.Request) {
	s.decide(w, r, verzoek.StatusDenied)
}

// decide applies an approval decision (the afnemer-side approval step). Only the
// beheerder role may approve/deny; the verstrekker that submitted the verzoek can
// therefore not self-approve its own request.
func (s *Service) decide(w http.ResponseWriter, r *http.Request, status verzoek.Status) {
	p, _ := principalFrom(r.Context())
	if !requireBeheerder(w, p) {
		return
	}

	id := r.PathValue("id")

	var body decideRequest
	_ = json.NewDecoder(r.Body).Decode(&body) // body is optional.

	prev, idx, err := s.store.Get(id)
	if err != nil {
		s.notFoundOrError(w, err)
		return
	}
	if prev.Status != verzoek.StatusPending {
		writeError(w, http.StatusConflict, "verzoek is not pending (status="+string(prev.Status)+")")
		return
	}

	// Record the deciding beheerder from the authenticated identity, falling back
	// to the optional body field only when the token carries no verstrekker id.
	decidedBy := p.Verstrekker
	if decidedBy == "" {
		decidedBy = body.DecidedBy
	}

	decidedAt := now().UTC()
	updated := prev
	updated.Status = status
	updated.DecidedAt = &decidedAt
	updated.DecidedBy = decidedBy
	updated.Reason = body.Reason

	if _, err := s.store.Update(prev, idx, updated); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.logAccess(r.Context(), "verzoek."+string(status), decidedBy, true, map[string]any{
		"verzoek_id":  id,
		"verstrekker": prev.Verstrekker,
		"decided_by":  decidedBy,
	})

	writeJSON(w, http.StatusOK, updated)
}

// Verwerking is a minimised projection of an ADL record returned in a result.
// Per the location rule it carries references and derived metadata, never a copy
// of the raw request/response payloads. Attribute source references already in
// the record (policies, information) are passed through unchanged.
type Verwerking struct {
	TraceID       string         `json:"trace_id"`
	SpanID        string         `json:"span_id"`
	ParentSpanID  string         `json:"parent_span_id,omitempty"`
	TransactionID string         `json:"transaction_id,omitempty"`
	EventName     string         `json:"event_name"`
	Timestamp     time.Time      `json:"timestamp"`
	Status        string         `json:"status"`
	Decision      *bool          `json:"decision,omitempty"`
	Afnemer       string         `json:"afnemer,omitempty"`
	Doel          string         `json:"doel,omitempty"`
	References    map[string]any `json:"references,omitempty"`
}

// handleResultaat serves GET /v1/inzicht/verzoeken/{id}/resultaat: it returns the
// requested processing activities, but only after the request has been approved and
// only to the verstrekker that submitted the verzoek.
//
// Grondslag / k-anonimiteit. The aggregate statistics endpoint applies a k-anonymity
// threshold because it is a statistical, non-purpose-bound disclosure: small buckets
// could single out individuals, so buckets with n<k are suppressed. This individual
// result path is different: retrieving it after an explicit, logged approval is a
// purpose-bound disclosure (doelgebonden verstrekking), not anonymisation. Its lawful
// basis is the approval decision itself, bound to the verzoek's doel. k-anonymity is
// therefore NOT applied here (it would defeat the very purpose of the inspection).
// Instead, data minimisation is enforced by SCOPING: the result is limited to the
// approved doel, the approved trace-id set, and the approved afnemer scope, and the
// records are projected to references + derived metadata (never raw payloads). The
// scope is bounded by what the beheerder approved, and disclosure only ever reaches
// the verstrekker that submitted (and thus owns) the verzoek.
func (s *Service) handleResultaat(w http.ResponseWriter, r *http.Request) {
	p, _ := principalFrom(r.Context())

	id := r.PathValue("id")

	v, _, err := s.store.Get(id)
	if err != nil {
		s.notFoundOrError(w, err)
		return
	}

	// Doelbinding: only the verstrekker that submitted the verzoek may read its
	// result. A caller with a different (or empty) verstrekker identity is refused.
	if p == nil || v.Verstrekker == "" || p.Verstrekker != v.Verstrekker {
		s.logAccess(r.Context(), "verzoek.resultaat", callerVerstrekker(r), false, map[string]any{
			"verzoek_id": id,
			"reason":     "not the submitting verstrekker",
		})
		writeError(w, http.StatusForbidden, "forbidden: not the verstrekker that submitted this verzoek")
		return
	}

	if v.Status != verzoek.StatusApproved {
		// Access to an unapproved (pending/denied) result is refused, and that
		// refusal is logged in the ADL as a denied access.
		s.logAccess(r.Context(), "verzoek.resultaat", v.Verstrekker, false, map[string]any{
			"verzoek_id": id,
			"status":     string(v.Status),
		})
		writeError(w, http.StatusForbidden, "verzoek is not approved (status="+string(v.Status)+")")
		return
	}

	f := query.Filter{
		Doel:      v.Doel,
		Afnemer:   v.Afnemer, // scope to the approved afnemer (data minimisation).
		TraceIDs:  v.TraceIDs,
		EventName: adl.EventAccessEvaluation,
	}
	if v.Vanaf != nil {
		f.From = *v.Vanaf
	}
	if v.Tot != nil {
		f.To = *v.Tot
	}

	records, err := s.source.Query(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	verwerkingen := make([]Verwerking, 0, len(records))
	for i := range records {
		verwerkingen = append(verwerkingen, projectRecord(&records[i]))
	}

	s.logAccess(r.Context(), "verzoek.resultaat", v.Verstrekker, true, map[string]any{
		"verzoek_id": id,
		"count":      len(verwerkingen),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"verzoek_id":   id,
		"verwerkingen": verwerkingen,
	})
}

// projectRecord minimises an ADL record into a Verwerking (references + derived
// metadata, no raw payload duplication).
func projectRecord(r *adl.Record) Verwerking {
	vw := Verwerking{
		TraceID:       r.TraceID,
		SpanID:        r.SpanID,
		ParentSpanID:  r.ParentSpanID,
		TransactionID: query.TransactionID(r),
		EventName:     r.EventName,
		Timestamp:     time.UnixMilli(int64(r.Timestamp)).UTC(),
		Status:        string(r.Status),
		Afnemer:       query.Afnemer(r),
		Doel:          query.Doel(r),
	}
	if d, ok := query.Decision(r); ok {
		vw.Decision = &d
	}
	// Pass through source references already present in attributes (policies,
	// information) - these are references, not raw payloads.
	if r.Attributes != nil {
		refs := map[string]any{}
		for _, k := range []string{adl.KeyPolicies, adl.KeyInformation, adl.KeyTransactionID} {
			if val, ok := r.Attributes[k]; ok {
				refs[k] = val
			}
		}
		if len(refs) > 0 {
			vw.References = refs
		}
	}
	return vw
}

func (s *Service) notFoundOrError(w http.ResponseWriter, err error) {
	if errors.Is(err, verzoek.ErrNotFound) {
		writeError(w, http.StatusNotFound, "verzoek not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}
