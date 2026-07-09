package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
)

// EventInzichtAccess is the ADL event_name used for access to the Inzicht API
// itself.
//
// The ADL standard defines five conformant event_name values, each mapping onto
// an AuthZEN API. Access to the Inzicht API is not a PDP access-evaluation, so
// adl.access_evaluation is not appropriate. The Inzicht operations enumerate a
// set of processing records (verwerkingen) matching a filter - a set-returning
// read over the log - which most closely matches the Resource Search API shape.
// We therefore record these accesses as adl.search_resource, the closest
// conformant value, so that access to the log is itself logged in the ADL (as
// the standard requires for access decisions on the log).
const EventInzichtAccess = adl.EventSearchResource

// logAccess writes one ADL record for an access to the Inzicht API. The subject
// is the requesting verstrekker, the action is the operation, and the response
// decision reflects whether the access was permitted. The record is appended to
// the same ADL write-ahead log used for decision logging.
func (s *Service) logAccess(ctx context.Context, op, verstrekker string, permit bool, reqContext map[string]any) {
	subjectID := verstrekker
	if subjectID == "" {
		subjectID = "unknown"
	}

	rec := &adl.Record{
		TraceID:   randomHex(16),
		SpanID:    randomHex(8),
		EventName: EventInzichtAccess,
		Timestamp: uint64(now().UnixMilli()),
		Status:    adl.StatusOk,
		Resource:  s.cfg.ADL.ResourceMap(map[string]any{"service.name": "inzicht"}),
		Body: map[string]any{
			adl.KeyRequest: map[string]any{
				"subject": map[string]any{"type": "verstrekker", "id": subjectID},
				"action":  map[string]any{"name": op},
				"context": reqContext,
			},
			adl.KeyResponse: map[string]any{"decision": permit},
		},
	}

	if err := s.adl.LogRecord(ctx, false, rec); err != nil && s.logger != nil {
		s.logger.Warn("inzicht: failed to log access", "op", op, "error", err)
	}
}

// randomHex returns n cryptographically-random bytes as lowercase hex.
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("inzicht: CSPRNG failure: " + err.Error())
	}
	return hex.EncodeToString(b)
}
