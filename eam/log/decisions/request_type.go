package decisions

import "github.com/goccy/go-json"

// AuthRequestType represents the type of request.
type AuthRequestType uint8

// List of supported request types.
const (
	EvaluationEndpoint AuthRequestType = iota + 1
	EvaluationsEndpoint
	SearchSubjectEndpoint
	SearchActionEndpoint
	SearchResourceEndpoint
)

// String implements the Stringer interface.
func (t AuthRequestType) String() string {
	switch t {
	case EvaluationEndpoint:
		return "evaluation"
	case EvaluationsEndpoint:
		return "evaluations"
	case SearchSubjectEndpoint:
		return "search_subject"
	case SearchActionEndpoint:
		return "search_action"
	case SearchResourceEndpoint:
		return "search_resource"
	default:
		return "???"
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (t AuthRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

var translate = map[string]AuthRequestType{
	"evaluation":      EvaluationEndpoint,
	"evaluations":     EvaluationsEndpoint,
	"search_subject":  SearchSubjectEndpoint,
	"search_action":   SearchActionEndpoint,
	"search_resource": SearchResourceEndpoint,
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *AuthRequestType) UnmarshalJSON(b []byte) error {
	*t = 0

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if t2, ok := translate[s]; ok {
		*t = t2
	}
	return nil
}
