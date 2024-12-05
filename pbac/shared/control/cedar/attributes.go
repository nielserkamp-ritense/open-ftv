package cedar

import (
	"log/slog"
	"maps"
	"strconv"
	"sync"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// NewAttributeBuilder returns the function prototype for building a new Cedar based attribute set.
func NewAttributeBuilder(logger *slog.Logger) models.AttributesBuilder {
	return func(in ...any) models.AttributeSet {
		return NewAttributeSet(logger, in...)
	}
}

// NewAttributeSet instantiates a new Cedar based attribute set.
func NewAttributeSet(logger *slog.Logger, in ...any) models.AttributeSet {
	a := &attributes{logger: logger, set: make(cedar.RecordMap)}
	for _, p := range in {
		switch t := p.(type) {
		case cedar.RecordMap:
			for k := range t {
				a.set[k] = t[k]
			}
		case models.AttributeSet:
			a.MergeAttributes(t)
		case models.Attribute:
			a.AddAttribute(t.Key, t.Value)
		case *models.Attribute:
			a.AddAttribute(t.Key, t.Value)
		}
	}
	return a
}

// AddAttribute implements the AttributeSet interface.
func (a *attributes) AddAttribute(key string, value any) {
	a.mutex.Lock()
	a.set[cedar.String(key)] = a.anyToValue(value)
	a.mutex.Unlock()
}

// GetAttribute implements the AttributeSet interface.
func (a *attributes) GetAttribute(key string) any {
	a.mutex.RLock()
	v := a.set[cedar.String(key)]
	a.mutex.RUnlock()
	return a.valueToAny(v)
}

// RemoveAttribute implements the AttributeSet interface.
func (a *attributes) RemoveAttribute(key string) {
	a.mutex.Lock()
	delete(a.set, cedar.String(key))
	a.mutex.Unlock()
}

// IterateAttributes implements the AttributeSet interface.
func (a *attributes) IterateAttributes(f models.AttributeIterator) {
	a.mutex.RLock()
	for k := range a.set {
		f(k.String(), a.valueToAny(a.set[k]))
	}
	a.mutex.RUnlock()
}

// MergeAttributes implements the AttributeSet interface.
func (a *attributes) MergeAttributes(in ...models.AttributeSet) {
	a.mutex.Lock()
	for i := range in {
		if set, ok := in[i].(*attributes); ok {
			set.mutex.RLock()
			maps.Copy(a.set, set.set)
			set.mutex.RUnlock()
		} else {
			in[i].IterateAttributes(func(key string, value any) {
				a.set[cedar.String(key)] = a.anyToValue(value)
			})
		}
	}
	a.mutex.Unlock()
}

func (a *attributes) valueToAny(in cedar.Value) any {
	switch t := in.(type) {
	case nil:
		return nil
	case cedar.String:
		return string(t)
	case cedar.Boolean:
		return bool(t)
	case cedar.Long:
		return int64(t)
	case cedar.Decimal:
		f, _ := strconv.ParseFloat(t.String(), 64)
		return f
	case cedar.Datetime:
		return time.UnixMilli(t.Milliseconds()).UTC()
	case cedar.Duration:
		return time.Millisecond * time.Duration(t.ToMilliseconds())

	case cedar.Set:
		return a.setToSlice(t)

	case cedar.Record:
		return a.recordToMap(t)

	default:
		a.logger.Warn("unsupported attribute conversion from Cedar", "type", t)
		return t.String()
	}
}

func (a *attributes) setToSlice(in cedar.Set) []any {
	s := in.Slice()

	out := make([]any, len(s))
	for i := range s {
		out[i] = a.valueToAny(s[i])
	}
	return out
}

func (a *attributes) recordToMap(in cedar.Record) map[string]any {
	m := in.Map()

	out := make(map[string]any, len(m))
	for k, v := range m {
		out[string(k)] = a.valueToAny(v)
	}
	return out
}

func (a *attributes) anyToValue(in any) cedar.Value {
	if cv, ok := in.(cedar.Value); ok {
		return cv
	}

	switch t := in.(type) {
	case nil:
		return nil
	case string:
		return cedar.String(t)
	case bool:
		return cedar.Boolean(t)
	case int:
		return cedar.Long(t)
	case int64:
		return cedar.Long(t)
	case float64:
		d, _ := cedar.NewDecimalFromFloat(t)
		return d
	case time.Time:
		return cedar.NewDatetime(t)
	case time.Duration:
		return cedar.NewDuration(t)

	case []string:
		s := make([]cedar.Value, len(t))
		for i := range t {
			s[i] = cedar.String(t[i])
		}
		return cedar.NewSet(s...)

	case []any:
		s := make([]cedar.Value, len(t))
		for i := range t {
			s[i] = a.anyToValue(t[i])
		}
		return cedar.NewSet(s...)

	case map[string]string:
		s := make(cedar.RecordMap, len(t))
		for k, v := range t {
			s[cedar.String(k)] = cedar.String(v)
		}
		return cedar.NewRecord(s)

	case map[string]any:
		s := make(cedar.RecordMap, len(t))
		for k, v := range t {
			s[cedar.String(k)] = a.anyToValue(v)
		}
		return cedar.NewRecord(s)

	default:
		a.logger.Warn("unsupported attribute conversion to Cedar", "type", t)
		return nil
	}
}

type attributes struct {
	logger *slog.Logger
	set    cedar.RecordMap
	mutex  sync.RWMutex
}
