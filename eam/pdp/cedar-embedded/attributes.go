package cedar_embedded

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func valueToAny(in cedar.Value) (any, error) {
	switch t := in.(type) {
	case nil:
		return nil, nil
	case cedar.String:
		return string(t), nil
	case cedar.Boolean:
		return bool(t), nil
	case cedar.Long:
		return int64(t), nil
	case cedar.Decimal:
		f, _ := strconv.ParseFloat(t.String(), 64)
		return f, nil
	case cedar.Datetime:
		return time.UnixMilli(t.Milliseconds()).UTC(), nil
	case cedar.Duration:
		return time.Millisecond * time.Duration(t.ToMilliseconds()), nil
	case cedar.Set:
		return setToSlice(t)
	case cedar.Record:
		return recordToMap(t)
	default:
		return t.String(), fmt.Errorf("unsupported attribute conversion from Cedar; type: [%T]", in)
	}
}

func setToSlice(in cedar.Set) ([]any, error) {
	s := in.Slice()
	out := make([]any, len(s))

	var err error
	for i := range s {
		if out[i], err = valueToAny(s[i]); err != nil {
			break
		}
	}

	return out, err
}

func recordToMap(in cedar.Record) (map[string]any, error) {
	m := in.Map()
	out := make(map[string]any, len(m))

	var err error
	for k, v := range m {
		if out[string(k)], err = valueToAny(v); err != nil {
			break
		}
	}

	return out, err
}

func anyToValue(in any) (cedar.Value, error) {
	if cv, ok := in.(cedar.Value); ok {
		return cv, nil
	}

	switch t := in.(type) {
	case nil:
		return nil, nil
	case string:
		return cedar.String(t), nil
	case bool:
		return cedar.Boolean(t), nil
	case int:
		return cedar.Long(t), nil
	case int64:
		return cedar.Long(t), nil
	case float64:
		return cedar.NewDecimalFromFloat(t)
	case time.Time:
		return cedar.NewDatetime(t), nil
	case time.Duration:
		return cedar.NewDuration(t), nil
	case []string:
		return stringSliceToValue(t)
	case []any:
		return anySliceToValue(t)
	case map[string]string:
		return stringsMapToValue(t)
	case map[string]any:
		return anyMapToValue(t)
	case *models.AttributeSet:
		return attributesToValue(t)
	default:
		return nil, fmt.Errorf("unsupported attribute conversion to Cedar; type [%T]", in)
	}
}

func stringSliceToValue(in []string) (cedar.Value, error) {
	s := make([]cedar.Value, len(in))
	for i := range in {
		s[i] = cedar.String(in[i])
	}
	return cedar.NewSet(s...), nil
}

func anySliceToValue(in []any) (cedar.Value, error) {
	s := make([]cedar.Value, len(in))
	for i := range in {
		v, err := anyToValue(in[i])
		if err != nil {
			return nil, err
		}
		s[i] = v
	}
	return cedar.NewSet(s...), nil
}

func stringsMapToValue(in map[string]string) (cedar.Value, error) {
	s := make(cedar.RecordMap, len(in))
	for k, v := range in {
		s[cedar.String(k)] = cedar.String(v)
	}
	return cedar.NewRecord(s), nil
}

func anyMapToValue(in map[string]any) (cedar.Value, error) {
	s := make(cedar.RecordMap, len(in))
	for k, v := range in {
		v2, err := anyToValue(v)
		if err != nil {
			return nil, err
		}
		s[cedar.String(k)] = v2
	}
	return cedar.NewRecord(s), nil
}

func attributesToValue(in *models.AttributeSet) (cedar.Value, error) {
	s := make(cedar.RecordMap)
	var err error
	in.IterateAttributes(func(attr *models.Attribute) {
		if err != nil {
			return
		}
		s[cedar.String(attr.Key())], err = anyToValue(attr.Value())
	})
	return cedar.NewRecord(s), err
}
