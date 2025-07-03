package schema

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/compare"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// Transformation represents a transformation of a field, another transformation or a value,
// or a list of two or more fields, transformations and/or values.
//
// The input field(s), transformation(s) and/or value(s) are used in the transformation process
// in the order assigned by the key of their corresponding map.
//
// E.g., if a field is compared (such as a greater or lesser comparison) to a transformation,
// the InputFields map will contain a field identifier with a key value of 1.
// And the InputTransforms map will contain a transformation identifier with a key value of 2.
//
// As an example, we want the following transformation: field_A <= field_B.
// The InputFields map will look like: {1: "field_A", 2: "field_B"}.
// The TransformationType will be "Compare", and the CompareType will be "IsLesserOrEqual".
//
// Another example could be to return a flag if a person is 18 years of age or older.
// In this case, the InputTransforms would contain an "Age" transformation with key 1,
// the CompareType would be "IsGreaterOrEqual",
// and the InputValues map would contain value 18 with key 2.
type Transformation struct {
	Object
	TransformationType enums.TransformationType
	CompareType        enums.CompareType
	Insensitive        bool
	ResultType         enums.FieldType
	IsPII              bool
	InputFields        map[int]string
	InputTransforms    map[int]string
	InputValues        map[int]any
	// hidden fields
	mutex      sync.Mutex
	transforms map[string]*Transformation
	rx         *regexp.Regexp
}

// FQID returns the fully qualified ID of this transformation.
func (t *Transformation) FQID() string {
	if t.parentTable != nil {
		return fmt.Sprintf("%s.%s", t.parentTable.ID, t.ID)
	}
	return t.ID
}

// GetTransformation returns the transformation definition for the given identifier.
func (t *Transformation) GetTransformation(id string) *Transformation {
	return t.transforms[id]
}

// GetRX returns the pre-compiled regular expression (if it exists).
func (t *Transformation) GetRX() *regexp.Regexp {
	return t.rx
}

// MarshalJSON implements the JSON Marshaler interface.
func (t *Transformation) MarshalJSON() ([]byte, error) {
	t2 := encodeTransformation{
		ID:                 t.ID,
		Description:        t.Description,
		TransformationType: t.TransformationType,
		ResultType:         t.ResultType,
		CompareType:        t.CompareType,
		IsPII:              t.IsPII,
		InputFields:        intStringToStringString(t.InputFields),
		InputTransforms:    intStringToStringString(t.InputTransforms),
		InputValues:        intAnyToStringAny(t.InputValues),
	}
	return json.Marshal(&t2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *Transformation) UnmarshalJSON(b []byte) error {
	var t2 encodeTransformation
	if err := json.Unmarshal(b, &t2); err != nil {
		return err
	}

	t.ID = t2.ID
	t.Description = t2.Description
	t.TransformationType = t2.TransformationType
	t.ResultType = t2.ResultType
	t.CompareType = t2.CompareType
	t.IsPII = t2.IsPII
	t.InputFields = stringStringToIntString(t2.InputFields)
	t.InputTransforms = stringStringToIntString(t2.InputTransforms)
	t.InputValues = stringAnyToIntAny(t2.InputValues)

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (t *Transformation) MarshalYAML() ([]byte, error) {
	f2 := encodeTransformation{
		ID:                 t.ID,
		Description:        t.Description,
		TransformationType: t.TransformationType,
		ResultType:         t.ResultType,
		CompareType:        t.CompareType,
		IsPII:              t.IsPII,
		InputFields:        intStringToStringString(t.InputFields),
		InputTransforms:    intStringToStringString(t.InputTransforms),
		InputValues:        intAnyToStringAny(t.InputValues),
	}
	return yaml.Marshal(&f2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *Transformation) UnmarshalYAML(b []byte) error {
	var t2 encodeTransformation
	if err := yaml.Unmarshal(b, &t2); err != nil {
		return err
	}

	t.ID = t2.ID
	t.Description = t2.Description
	t.TransformationType = t2.TransformationType
	t.ResultType = t2.ResultType
	t.CompareType = t2.CompareType
	t.IsPII = t2.IsPII
	t.InputFields = stringStringToIntString(t2.InputFields)
	t.InputTransforms = stringStringToIntString(t2.InputTransforms)
	t.InputValues = stringAnyToIntAny(t2.InputValues)

	return nil
}

func intStringToStringString(in map[int]string) map[string]string {
	if in == nil {
		return nil
	}

	out := make(map[string]string, len(in))
	for k, v := range in {
		out[strconv.Itoa(k)] = v
	}
	return out
}

func intAnyToStringAny(in map[int]any) map[string]any {
	if in == nil {
		return nil
	}

	out := make(map[string]any, len(in))
	for k, v := range in {
		out[strconv.Itoa(k)] = v
	}
	return out
}

func stringStringToIntString(in map[string]string) map[int]string {
	if in == nil {
		return nil
	}

	out := make(map[int]string, len(in))
	for k, v := range in {
		out[int(convert.AnyToInt64(k))] = v
	}
	return out
}

func stringAnyToIntAny(in map[string]any) map[int]any {
	if in == nil {
		return nil
	}

	out := make(map[int]any, len(in))
	for k, v := range in {
		out[int(convert.AnyToInt64(k))] = v
	}
	return out
}

type encodeTransformation struct {
	ID                 string                   `json:"id" yaml:"id"`
	Description        string                   `json:"description,omitempty" yaml:"description,omitempty"`
	TransformationType enums.TransformationType `json:"transformationType" yaml:"transformationType"`
	ResultType         enums.FieldType          `json:"resultType,omitempty" yaml:"resultType,omitempty"`
	CompareType        enums.CompareType        `json:"compareType,omitempty" yaml:"compareType,omitempty"`
	IsPII              bool                     `json:"isPII,omitempty" yaml:"isPII,omitempty"`
	InputFields        map[string]string        `json:"inputFields,omitempty" yaml:"inputFields,omitempty"`
	InputTransforms    map[string]string        `json:"inputTransformations,omitempty" yaml:"inputTransformations,omitempty"`
	InputValues        map[string]any           `json:"inputValues,omitempty" yaml:"inputValues,omitempty"`
}

// Fix (re)sets the parent-child relationships for this object.
func (t *Transformation) Fix(table *Table) {
	t.mutex.Lock()
	t.fix(table)
	t.mutex.Unlock()
}

func (t *Transformation) fix(table *Table) {
	if table != nil {
		t.Object.fix(&table.Parent)
		t.Object.parentTable = table

		t.Object.fields = make(map[string]*Field, len(t.InputFields))
		for i := range t.InputFields {
			id := t.InputFields[i]
			if f := table.fields[id]; f != nil {
				t.fields[id] = f
			}
		}

		t.transforms = make(map[string]*Transformation, len(t.InputTransforms))
		for i := range t.InputTransforms {
			id := t.InputTransforms[i]
			if transform := table.transforms[id]; transform != nil {
				t.transforms[id] = transform
			}
		}

		switch t.CompareType {
		case enums.IsLike, enums.IsNotLike:
			t.rx, _ = compare.RXFromLike(t.InputValues[2])
		case enums.MatchRegex, enums.NotMatchRegex:
			t.rx, _ = compare.RXFromString(t.InputValues[2])
		default:
		}
	}
}
