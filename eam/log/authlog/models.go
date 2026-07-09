package authlog

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// AuthRecord represents the details to be stored in the authorisation log.
type AuthRecord struct {
	ClientIP        string              `json:"clientAddress,omitempty"`
	RequestTime     *time.Time          `json:"requestTime,omitempty"`
	RvvaID          string              `json:"rvvaId,omitempty"`
	Principal       models.Entity       `json:"principal,omitempty"`
	Action          models.Entity       `json:"action,omitempty"`
	Resource        models.Entity       `json:"resource,omitempty"`
	Decision        bool                `json:"decision"`
	DecisionContext models.AttributeSet `json:"decisionContext,omitempty"`
	TraceParent     string              `json:"traceparent,omitempty"` // https://www.w3.org/TR/trace-context/
	TraceState      string              `json:"tracestate,omitempty"`  // https://www.w3.org/TR/trace-context/

	// The following fields carry extra context for ADL-conformant logging (eam/log/adl).
	// They are excluded from JSON so the existing OpenSearch record shape is unchanged.

	// EventName is the ADL event name (one of adl.*). Empty defaults to adl.access_evaluation.
	EventName string `json:"-"`
	// Errored reports that the PDP could not evaluate the request (ADL status Error). A
	// denial is NOT an error and MUST leave this false.
	Errored bool `json:"-"`
	// RequestContext is the request's context AttributeSet, used to reconstruct the full
	// AuthZEN request in the ADL record body.
	RequestContext models.AttributeSet `json:"-"`

	// FSCTransactionID is the FSC TransactionID (from the Fsc-Transaction-Id HTTP header)
	// when the decision request crossed an FSC inway/outway. It populates
	// attributes[adl.fsc.transaction_id]; it MUST be set on the FSC path and MUST be empty
	// otherwise (e.g. a pure AuthZEN PDP request carries no FSC transaction).
	FSCTransactionID string `json:"-"`
}
