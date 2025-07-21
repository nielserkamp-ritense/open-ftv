package authlog

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// AuthRecord represents the details to be stored in the authorization-log.
type AuthRecord struct {
	ClientIP        string               `json:"clientAddress,omitempty"`
	RequestTime     *time.Time           `json:"requestTime,omitempty"`
	RvvaID          string               `json:"rvvaId,omitempty"`
	Principal       *models.Entity       `json:"principal,omitempty"`
	Action          *models.Entity       `json:"action,omitempty"`
	Resource        *models.Entity       `json:"resource,omitempty"`
	Decision        bool                 `json:"decision"`
	DecisionContext *models.AttributeSet `json:"decisionContext,omitempty"`
	TraceParent     string               `json:"traceparent,omitempty"` // https://www.w3.org/TR/trace-context/
	TraceState      string               `json:"tracestate,omitempty"`  // https://www.w3.org/TR/trace-context/
}
