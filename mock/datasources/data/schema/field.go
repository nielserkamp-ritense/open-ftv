package schema

import (
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Field represents a field in a datasource table.
type Field struct {
	Object
	Type          enums.FieldType
	IsArray       bool
	IsEnum        bool
	IsPII         bool
	MinLen        int
	MaxLen        int
	Format        string
	MinValue      any
	MaxValue      any
	AllowedValues []any
	// hidden fields
	mutex sync.Mutex
}

// FQID returns the fully qualified ID of this field.
func (f *Field) FQID() string {
	if f.parentTable != nil {
		return fmt.Sprintf("%s.%s", f.parentTable.ID, f.ID)
	}
	return f.ID
}

// ConvertValue converts the given value in accordance with the field type.
func (f *Field) ConvertValue(v any) any {
	if v == nil {
		return v
	}

	switch f.Type {
	case enums.IntegerType:
		return convert.AnyToInt64(v)
	case enums.UnsignedIntegerType:
		return convert.AnyToUint64(v)
	case enums.FloatType:
		return convert.AnyToFloat64(v)
	case enums.BooleanType:
		return convert.AnyToBool(v)
	case enums.DateType, enums.TimeType, enums.DateTimeType:
		return convert.AnyToDateTime(v)
	default:
		return convert.AnyToString(v)
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (f *Field) MarshalJSON() ([]byte, error) {
	f2 := encodeField{
		ID:            f.ID,
		Description:   f.Description,
		Type:          f.Type,
		IsArray:       f.IsArray,
		IsEnum:        f.IsEnum,
		IsPII:         f.IsPII,
		MinLen:        f.MinLen,
		MaxLen:        f.MaxLen,
		Format:        f.Format,
		MinValue:      f.MinValue,
		MaxValue:      f.MaxValue,
		AllowedValues: f.AllowedValues,
	}
	return json.Marshal(&f2)
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (f *Field) UnmarshalJSON(b []byte) error {
	var f2 encodeField
	if err := json.Unmarshal(b, &f2); err != nil {
		return err
	}

	f.ID = f2.ID
	f.Description = f2.Description
	f.Type = f2.Type
	f.IsArray = f2.IsArray
	f.IsEnum = f2.IsEnum
	f.IsPII = f2.IsPII
	f.MinLen = f2.MinLen
	f.MaxLen = f2.MaxLen
	f.Format = f2.Format
	f.MinValue = f2.MinValue
	f.MaxValue = f2.MaxValue
	f.AllowedValues = f2.AllowedValues

	return nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (f *Field) MarshalYAML() ([]byte, error) {
	f2 := encodeField{
		ID:            f.ID,
		Description:   f.Description,
		Type:          f.Type,
		IsArray:       f.IsArray,
		IsEnum:        f.IsEnum,
		IsPII:         f.IsPII,
		MinLen:        f.MinLen,
		MaxLen:        f.MaxLen,
		Format:        f.Format,
		MinValue:      f.MinValue,
		MaxValue:      f.MaxValue,
		AllowedValues: f.AllowedValues,
	}
	return yaml.Marshal(&f2)
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (f *Field) UnmarshalYAML(b []byte) error {
	var f2 encodeField
	if err := yaml.Unmarshal(b, &f2); err != nil {
		return err
	}

	f.ID = f2.ID
	f.Description = f2.Description
	f.Type = f2.Type
	f.IsArray = f2.IsArray
	f.IsEnum = f2.IsEnum
	f.IsPII = f2.IsPII
	f.MinLen = f2.MinLen
	f.MaxLen = f2.MaxLen
	f.Format = f2.Format
	f.MinValue = f2.MinValue
	f.MaxValue = f2.MaxValue
	f.AllowedValues = f2.AllowedValues

	return nil
}

type encodeField struct {
	ID            string          `json:"id" yaml:"id"`
	Description   string          `json:"description,omitempty" yaml:"description,omitempty"`
	Type          enums.FieldType `json:"type" yaml:"type"`
	IsArray       bool            `json:"isArray,omitempty" yaml:"isArray,omitempty"`
	IsEnum        bool            `json:"isEnum,omitempty" yaml:"isEnum,omitempty"`
	IsPII         bool            `json:"isPII,omitempty" yaml:"isPII,omitempty"`
	MinLen        int             `json:"minLen,omitempty" yaml:"minLen,omitempty"`
	MaxLen        int             `json:"maxLen,omitempty" yaml:"maxLen,omitempty"`
	Format        string          `json:"format,omitempty" yaml:"format,omitempty"`
	MinValue      any             `json:"minValue,omitempty" yaml:"minValue,omitempty"`
	MaxValue      any             `json:"maxValue,omitempty" yaml:"maxValue,omitempty"`
	AllowedValues []any           `json:"allowedValues,omitempty" yaml:"allowedValues,omitempty"`
}

// Fix (re)sets the parent-child relationships for this field.
func (f *Field) Fix(table *Table, field *Field) {
	f.mutex.Lock()
	f.fix(table, field)
	f.mutex.Unlock()
}

func (f *Field) fix(table *Table, field *Field) {
	if table != nil {
		f.parentTable = table
		f.Object.Fix(&table.Parent)
	}
	if field != nil {
		f.parentField = field
		f.Object.Fix(&field.Parent)
	}

	f.Object.FixFields(nil, f) // fix child fields (if any).
}
