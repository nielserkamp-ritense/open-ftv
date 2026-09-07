package search

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
)

const (
	defaultLimit = 100
	maxLimit     = 1000
	maxStringLen = 100
	maxBundles   = 10
	maxPeriod    = 7 * 24 * time.Hour
	traceIDBytes = 16 // W3C trace-id: 16 bytes (32 hex characters).
	spanIDBytes  = 8  // W3C span-id: 8 bytes (16 hex characters).
)

// Criteria contains the parameter to use for a search.
//
// Note that the use of SubjectType, SubjectId, ActionName, ResourceType and/or ResourceId
// may considerably impact performance.
type Criteria struct {
	Limit        int                         // default 100; must be less than 1000.
	ID           int64                       // selects a single record by identifier (if it exists); mutually exclusive with all other criteria.
	From         time.Time                   // not more than 5 years ago; mutually exclusive with Recent.
	To           time.Time                   // default is current timestamp + 1 second; not more than 7 days after From; mutually exclusive with ID.
	Recent       time.Duration               // defines selection period upto current timestamp; mutually exclusive with From.
	RequestTypes []decisions.AuthRequestType // must be valid request types; duplicates not allowed.
	Bundles      []int64                     // must be positive integers; duplicates not allowed; max 10 items.
	TraceId      string                      // must be 32 characters hex.
	SpanId       string                      // must be 16 characters hex.
	SubjectType  string                      // not more than 100 characters.
	SubjectId    string                      // not more than 100 characters.
	ActionName   string                      // not more than 100 characters.
	ResourceType string                      // not more than 100 characters.
	ResourceId   string                      // not more than 100 characters.
}

// Test checks if the criteria are valid.
//
// It returns nil if all criteria are valid or a concatenation of the errors found.
func (c *Criteria) Test() error {
	var errs []error
	var err error

	if c.ID > 0 {
		if !c.From.IsZero() || !c.To.IsZero() || c.Recent != 0 {
			errs = append(errs, fmt.Errorf("id search is mutually exclusive with period selection: %s, %s, %s", c.From.String(), c.To.String(), c.Recent.String()))
		}
	} else {
		if c.TraceId != "" && c.From.IsZero() && c.To.IsZero() && c.Recent == 0 {
			// The +1s on To (and so on From, exactly 5 years before it) is a
			// buffer against the "less than 5 years ago" check below calling
			// time.Now() again a moment later: without it, that later call
			// would occasionally push From just past the 5-year cutoff.
			c.To = time.Now().UTC().Add(time.Second)
			c.From = c.To.AddDate(-5, 0, 0)
		}
		if c.To.IsZero() {
			c.To = time.Now().UTC().Add(time.Second)
		}
		if c.Recent > 0 {
			if c.From.IsZero() {
				c.From = c.To.Add(-c.Recent)
			} else {
				errs = append(errs, errors.New("recent period and from timestamp are mutually exclusive"))
			}
		}

		switch {
		case c.From.IsZero():
			errs = append(errs, errors.New("from timestamp must be filled"))
		case c.From.Before(time.Now().AddDate(-5, 0, 0)):
			errs = append(errs, errors.New("selection period must be less than 5 years ago"))
		case c.From.Add(maxPeriod).Before(c.To) && c.TraceId == "":
			errs = append(errs, fmt.Errorf("selection period cannot be larger than %s", maxPeriod.String()))
		}
	}
	slices.Sort(c.RequestTypes)

	for i := range c.RequestTypes {
		switch c.RequestTypes[i] {
		case decisions.EvaluationEndpoint, decisions.EvaluationsEndpoint,
			decisions.SearchSubjectEndpoint, decisions.SearchActionEndpoint, decisions.SearchResourceEndpoint:
		// good
		default:
			errs = append(errs, fmt.Errorf("invalid request type: %d", c.RequestTypes[i]))
		}

		if i > 0 && c.RequestTypes[i] == c.RequestTypes[i-1] {
			errs = append(errs, fmt.Errorf("duplicate request type: %d", c.RequestTypes[i]))
		}
	}

	if len(c.Bundles) > maxBundles {
		errs = append(errs, fmt.Errorf("bundles exceeds maximum (%d): %d", maxBundles, len(c.Bundles)))
	} else {
		slices.Sort(c.Bundles)

		for i := range c.Bundles {
			if i > 0 && c.Bundles[i] == c.Bundles[i-1] {
				errs = append(errs, fmt.Errorf("duplicate request type: %d", c.RequestTypes[i]))
			}
		}
	}

	if c.TraceId != "" {
		traceID, decErr := hex.DecodeString(c.TraceId)
		if decErr != nil {
			errs = append(errs, fmt.Errorf("failed to parse trace id: %w", decErr))
		}

		if len(traceID) != traceIDBytes {
			errs = append(errs, fmt.Errorf("trace id must be 16 bytes (32 hex characters) long"))
		}
	}

	if c.SpanId != "" {
		spanID, decErr := hex.DecodeString(c.SpanId)
		if decErr != nil {
			errs = append(errs, fmt.Errorf("failed to parse span id: %w", decErr))
		}

		if len(spanID) != spanIDBytes {
			errs = append(errs, fmt.Errorf("span id must be 8 bytes (16 hex characters) long"))
		}
	}

	if len(c.SubjectType) > maxStringLen {
		errs = append(errs, fmt.Errorf("subject type too long (%d > %d)", len(c.SubjectType), maxStringLen))
	}
	if len(c.SubjectId) > maxStringLen {
		errs = append(errs, fmt.Errorf("subject id too long (%d > %d)", len(c.SubjectId), maxStringLen))
	}
	if len(c.ActionName) > maxStringLen {
		errs = append(errs, fmt.Errorf("action name too long (%d > %d)", len(c.ActionName), maxStringLen))
	}
	if len(c.ResourceType) > maxStringLen {
		errs = append(errs, fmt.Errorf("resource type too long (%d > %d)", len(c.ResourceType), maxStringLen))
	}
	if len(c.ResourceId) > maxStringLen {
		errs = append(errs, fmt.Errorf("resource id too long (%d > %d)", len(c.ResourceId), maxStringLen))
	}

	switch {
	case c.Limit == 0:
		c.Limit = defaultLimit
	case c.Limit < 0:
		errs = append(errs, fmt.Errorf("limit must be positive number: %d", c.Limit))
	case c.Limit > maxLimit:
		errs = append(errs, fmt.Errorf("limit exceeds maximum (%d): %d", maxLimit, c.Limit))
	}

	if err = errors.Join(errs...); err != nil {
		return &ParameterError{msg: err.Error()}
	}
	return nil
}
