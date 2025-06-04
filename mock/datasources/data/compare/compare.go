package compare

import (
	"regexp"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Params represents the parameters for the comparison of a data value,
// against another value, a list of other values, a regular expression, or if it exists or not.
type Params struct {
	Type        enums.FieldType   // the type of input field.
	Compare     enums.CompareType // the type of comparison.
	Insensitive bool              // force case-insensitive string data comparisons.
	Input       any               // the input field to compare.
	Value       any               // a single value to compare against.
	Values      []any             // a set of values to compare against.
	RX          *regexp.Regexp    // a regular expression to compare against.
}

// Compare performs a comparison based on the given parameters.
//
// See Params for an explanation on the various input parameters.
//
// Unsupported comparison types will return a negative outcome.
func Compare(params Params) bool {
	switch params.Compare {
	case enums.NotExists:
		return params.Input == nil
	case enums.Exists:
		return params.Input != nil
	case enums.IsEqual:
		return params.equal(params.Value)
	case enums.IsNotEqual:
		return !params.equal(params.Value)
	case enums.IsLesser:
		return params.lesser()
	case enums.IsLesserOrEqual:
		return !params.greater()
	case enums.IsGreater:
		return params.greater()
	case enums.IsGreaterOrEqual:
		return !params.lesser()
	case enums.InList:
		return params.inList()
	case enums.NotInList:
		return !params.inList()
	case enums.IsLike, enums.MatchRegex:
		return params.matchRX()
	case enums.IsNotLike, enums.NotMatchRegex:
		return !params.matchRX()
	default:
		return false
	}
}

func (params *Params) equal(v2 any) bool {
	return Equal(params.Type, params.Insensitive, params.Input, v2)
}

func (params *Params) greater() bool {
	return Greater(params.Type, params.Insensitive, params.Input, params.Value)
}

func (params *Params) lesser() bool {
	return Lesser(params.Type, params.Insensitive, params.Input, params.Value)
}

func (params *Params) inList() bool {
	for i := range params.Values {
		if ok := params.equal(params.Values[i]); ok {
			return true
		}
	}
	return false
}

func (params *Params) matchRX() bool {
	return params.RX.MatchString(convert.AnyToString(params.Input))
}
