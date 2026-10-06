package csv

import "reflect"

func toSlice[T any](v any, convert func(any) T) []T {
	switch t := v.(type) {
	case nil:
		return []T{}
	case []T:
		return t
	case T:
		return []T{t}
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return []T{convert(v)}
	}

	out := make([]T, rv.Len())
	for i := range out {
		out[i] = convert(rv.Index(i).Interface())
	}

	return out
}
