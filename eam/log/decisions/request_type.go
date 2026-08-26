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

// unknownRequestTypeName is returned by String for a request type with no entry in requestTypeInfos.
const unknownRequestTypeName = "???"

// JSON names and Logius ADL event names for each request type.
const (
	nameEvaluation     = "evaluation"
	nameEvaluations    = "evaluations"
	nameSearchSubject  = "search_subject"
	nameSearchAction   = "search_action"
	nameSearchResource = "search_resource"

	eventNameAccessEvaluation  = "adl.access_evaluation"
	eventNameAccessEvaluations = "adl.access_evaluations"
	eventNameSearchSubject     = "adl.search_subject"
	eventNameSearchAction      = "adl.search_action"
	eventNameSearchResource    = "adl.search_resource"
)

// requestTypeInfo pairs a request type's JSON name with its Logius ADL event_name in one table.
type requestTypeInfo struct {
	name      string
	eventName string
}

var requestTypeInfos = map[AuthRequestType]requestTypeInfo{
	EvaluationEndpoint:     {name: nameEvaluation, eventName: eventNameAccessEvaluation},
	EvaluationsEndpoint:    {name: nameEvaluations, eventName: eventNameAccessEvaluations},
	SearchSubjectEndpoint:  {name: nameSearchSubject, eventName: eventNameSearchSubject},
	SearchActionEndpoint:   {name: nameSearchAction, eventName: eventNameSearchAction},
	SearchResourceEndpoint: {name: nameSearchResource, eventName: eventNameSearchResource},
}

var requestTypeByName = func() map[string]AuthRequestType {
	m := make(map[string]AuthRequestType, len(requestTypeInfos))
	for t, info := range requestTypeInfos {
		m[info.name] = t
	}

	return m
}()

// String implements the Stringer interface.
func (t AuthRequestType) String() string {
	if info, ok := requestTypeInfos[t]; ok {
		return info.name
	}

	return unknownRequestTypeName
}

// MarshalJSON implements the json.Marshaler interface.
func (t AuthRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// EventName returns the Logius ADL event_name for the request type.
func (t AuthRequestType) EventName() string {
	return requestTypeInfos[t].eventName
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *AuthRequestType) UnmarshalJSON(b []byte) error {
	*t = 0

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if t2, ok := requestTypeByName[s]; ok {
		*t = t2
	}

	return nil
}
