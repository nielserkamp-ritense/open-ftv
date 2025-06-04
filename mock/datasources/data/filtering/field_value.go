package filtering

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// FieldValueFilter represents the details of a filter operation.
//
// Try not to instantiate FieldValueFilter by hand as it may not work as you expect.
// Consider using the Exists(), Compare() and List() functions to instantiate a FieldValueFilter for a specific operation.
// Take good care when configuring YAML or JSON files used to read FieldValueFilter from.
//
// By default, a filter executes on the primary table.
// Use the OnJoin() or OnTable() functions to change this behavior.
// OnJoin() executes the filter only for the join it is linked to.
// OnTable() executes the filter every time the result data is selected from the linked table.
//
// By default, string comparisons are done case-sensitive.
// Use the CaseInsensitive() function to change string comparisons to be case-insensitive.
//
// Use the Prepare() function to prepare a FieldValueFilter for use with a specific datasource and set of joins (if any).
//
// Once the filter is prepared, you can use the MatchOnPrimaryData() and MatchOnJoinData() functions to filter your result data.
type FieldValueFilter struct {
	Level       enums.FilterLevel `json:"level" yaml:"level"`                                         // the level at which the filter must be applied.
	Insensitive bool              `json:"caseInsensitive,omitempty" yaml:"caseInsensitive,omitempty"` // indicates case insensitivity when comparing strings.
	Join        string            `json:"join,omitempty" yaml:"join,omitempty"`                       // the specific join for which the filter must be applied (mutually exclusive with table).
	Table       string            `json:"table,omitempty" yaml:"table,omitempty"`                     // the specific table for which the filter must be applied (mutually exclusive with join).
	Field       string            `json:"field" yaml:"field"`                                         // the field for which the filter must be applied.
	Compare     enums.CompareType `json:"compare,omitempty" yaml:"compare,omitempty"`                 // the type of comparison for the filter.
	Value       any               `json:"value,omitempty" yaml:"value,omitempty"`                     // the value to compare against (mutually exclusive with values).
	Values      []any             `json:"values,omitempty" yaml:"values,omitempty"`                   // the values to compare against (mutually exclusive with value).
	// hidden fields
	table *schema.Table
	join  *schema.Join
	field *schema.Field
	rx    *regexp.Regexp
	list  map[string]struct{}
}

// Exists returns a filter based on whether the field has a value or not.
//
// By default, the filter executes on the primary table.
// Use the OnTable() or OnJoin() functions to change this behavior.
func Exists(field string, exists bool) *FieldValueFilter {
	if exists {
		return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: enums.Exists}
	}
	return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: enums.NotExists}
}

// Compare returns a filter based on comparison with a single value.
//
// The function "fixes" certain unusual inputs.
// See the code for more details.
//
// By default, the filter executes on the primary table.
// Use the OnTable() or OnJoin() functions to change this behavior.
func Compare(field string, cmp enums.CompareType, value any) *FieldValueFilter {
	// make sure "unusual" comparison types return a valid filter.
	switch cmp {
	case enums.Exists, enums.NotExists:
		// ignore the value.
		return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: cmp}
	case enums.InList:
		// special case with a single value.
		return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: enums.IsEqual, Value: value}
	case enums.NotInList:
		// special case with a single value.
		return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: enums.IsNotEqual, Value: value}
	default:
		return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: cmp, Value: value}
	}
}

// List returns a filter based on comparison with a list of values.
//
// By default, the filter executes on the primary table.
// Use the OnTable() or OnJoin() functions to change this behavior.
func List(field string, exists bool, values ...any) *FieldValueFilter {
	if exists {
		return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: enums.InList, Values: values}
	}
	return &FieldValueFilter{Level: enums.PrimaryLevel, Field: field, Compare: enums.NotInList, Values: values}
}

// OnTable forces the filter to execute on the given table.
func (f *FieldValueFilter) OnTable(table string) *FieldValueFilter {
	f.Level = enums.AnyLevel
	f.Table = table
	return f
}

// OnJoin forces the filter to execute on the given join.
func (f *FieldValueFilter) OnJoin(join string) *FieldValueFilter {
	f.Level = enums.JoinLevel
	f.Join = join
	return f
}

// CaseInsensitive forces the filter to execute string comparisons case-insensitive.
func (f *FieldValueFilter) CaseInsensitive() *FieldValueFilter {
	f.Insensitive = true
	return f
}

// CaseSensitive forces the filter to execute string comparisons case-sensitive.
//
// This is also the default.
func (f *FieldValueFilter) CaseSensitive() *FieldValueFilter {
	f.Insensitive = false
	return f
}

// Prepare implements the Filterer interface.
//
// This function prepares the filter for operation on the given datasource and joins.
//
// The datasource allows the filter to find specific tables.
// The joins allow the filter to find specific joins.
func (f *FieldValueFilter) Prepare(ds *schema.Datasource, joins []*schema.Join) error {
	if err := f.checkConsistency(); err != nil {
		return err
	}

	if err := f.checkCompare(); err != nil {
		return err
	}

	if f.Table != "" || f.Level == enums.AnyLevel {
		if err := f.checkTable(ds); err != nil {
			return err
		}
	}

	if f.Join != "" || f.Level == enums.JoinLevel {
		if err := f.checkJoin(joins); err != nil {
			return err
		}
	}

	if f.field == nil {
		return f.checkField(ds)
	}
	return nil
}

func (f *FieldValueFilter) checkConsistency() error {
	if f.Table != "" && f.Join != "" {
		return fmt.Errorf("filtering on a specific table or join is mutually exclusive")
	}

	if f.Value != nil && f.Values != nil {
		return fmt.Errorf("filtering on a single or multiple values is mutually exclusive")
	}

	return nil
}

func (f *FieldValueFilter) checkCompare() error {
	switch f.Compare {
	case enums.Exists, enums.NotExists:
		if f.Value != nil || f.Values != nil {
			return fmt.Errorf("value and values must not be present")
		}

	case enums.InList, enums.NotInList:
		if f.Values == nil {
			return fmt.Errorf("values must be present for list comparisons")
		}
		return f.fixList()

	case enums.IsEqual, enums.IsNotEqual, enums.IsLesser, enums.IsLesserOrEqual, enums.IsGreater, enums.IsGreaterOrEqual:
		if f.Value == nil {
			return fmt.Errorf("value must be present for compound comparisons")
		}

	case enums.IsLike, enums.IsNotLike:
		if f.Value == nil {
			return fmt.Errorf("value must be present for like comparisons")
		}
		return f.fixLike()

	case enums.MatchRegex, enums.NotMatchRegex:
		if f.Value == nil {
			return fmt.Errorf("value must be present for regular expression comparisons")
		}
		return f.fixRX()

	default:
		return fmt.Errorf("invalid comparison type")
	}

	return nil
}

func (f *FieldValueFilter) checkTable(ds *schema.Datasource) error {
	if f.Level != enums.AnyLevel {
		return fmt.Errorf("filter on table must match at any level")
	}

	f.table = ds.Table(f.Table)
	if f.table == nil {
		return fmt.Errorf("table [%s] not found", f.Table)
	}

	f.field = f.table.Field(f.Field)
	if f.field == nil {
		return fmt.Errorf("field [%s] for table [%s] not found", f.Field, f.Table)
	}

	return nil
}

func (f *FieldValueFilter) checkJoin(joins []*schema.Join) error {
	if f.Level != enums.JoinLevel {
		return fmt.Errorf("filter on join must match at join level")
	}

	f.join = nil
	for _, join := range joins {
		if strings.EqualFold(join.GetJoinID(), f.Join) {
			f.join = join
			break
		}
	}
	if f.join == nil {
		return fmt.Errorf("join [%s] not found", f.Join)
	}

	src := f.join.GetSource()
	if src == nil {
		return fmt.Errorf("source table for join [%s] not found", f.Join)
	}

	f.field = src.Field(f.Field)
	if f.field == nil {
		return fmt.Errorf("field [%s] for join [%s] not found", f.Field, f.Join)
	}

	f.table = src
	return nil
}

func (f *FieldValueFilter) checkField(ds *schema.Datasource) error {
	parts := strings.Split(f.Field, ".")
	switch len(parts) {
	case 1:
		for _, t := range ds.Tables {
			if field := t.Field(f.Field); field != nil {
				if f.field != nil {
					return fmt.Errorf("field [%s] not unique", f.Field)
				}
				f.table = t
				f.field = field
			}
		}

	default:
		f.table = ds.Table(parts[0])
		if f.table == nil {
			return fmt.Errorf("table [%s] not found", parts[0])
		}
		f.field = f.table.Field(parts[1])
	}

	if f.field == nil {
		return fmt.Errorf("field [%s] not found", f.Field)
	}
	return nil
}

func (f *FieldValueFilter) fixList() error {
	f.list = make(map[string]struct{}, len(f.Values))
	for i := range f.Values {
		f.list[fmt.Sprintf("%v", f.Values[i])] = struct{}{}
	}
	return nil
}

func (f *FieldValueFilter) fixLike() error {
	return f.buildRX(likeReplacer.Replace(convert.AnyToString(f.Value)))
}

func (f *FieldValueFilter) fixRX() error {
	return f.buildRX(convert.AnyToString(f.Value))
}

func (f *FieldValueFilter) buildRX(s string) (err error) {
	if !strings.HasPrefix(s, "^") {
		s = fmt.Sprintf("^%s", s)
	}
	if !strings.HasSuffix(s, "$") {
		s = fmt.Sprintf("%s$", s)
	}

	f.rx, err = regexp.Compile(s)
	return
}

var likeReplacer = strings.NewReplacer(".", "\\.", "+", "\\+", "(", "\\(", ")", "\\)", "[", "\\[", "]", "\\]", "{", "\\{", "}", "\\}", "%", ".*", "*", ".*", "?", ".")
