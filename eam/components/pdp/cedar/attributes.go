package cedar

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// NewAttributeBuilder returns the function prototype for building a new Cedar based attribute set.
func NewAttributeBuilder(logger *slog.Logger) models.AttributesBuilder {
	return func(in ...any) models.AttributeSet {
		return NewAttributeSet(logger, in...)
	}
}

// NewAttributeSet instantiates a new Cedar based attribute set.
func NewAttributeSet(logger *slog.Logger, in ...any) models.AttributeSet {
	a := &attributes{
		logger:   logger,
		set:      models.NewAttributeSet(),
		cedarSet: make(cedar.RecordMap),
	}

	for _, p := range in {
		switch t := p.(type) {
		case cedar.Record:
			for k, v := range t.Map() {
				k2, v2 := string(k), a.valueToAny(v)
				a.addOriginalAttribute(k2, v2, v2, "")
			}
		case cedar.RecordMap:
			for k, v := range t {
				k2, v2 := string(k), a.valueToAny(v)
				a.addOriginalAttribute(k2, v2, v2, "")
			}
		case models.AttributeSet:
			a.MergeAttributes(t)
		case models.Attribute:
			a.addOriginalAttribute(t.Key(), t.Value(), t.Original(), t.Type())
		}
	}
	return a
}

// AddAttribute implements the AttributeSet interface.
func (a *attributes) AddAttribute(key string, value any) {
	a.AddOriginalAttribute(key, value, value, "")
}

// AddAttributeWithType implements the AttributeSet interface.
func (a *attributes) AddAttributeWithType(key string, value any, tp string) {
	a.AddOriginalAttribute(key, value, value, tp)
}

// AddOriginalAttribute implements the AttributeSet interface.
func (a *attributes) AddOriginalAttribute(key string, value, original any, tp string) {
	a.mutex.Lock()
	a.addOriginalAttribute(key, value, original, tp)
	a.mutex.Unlock()
}

func (a *attributes) addOriginalAttribute(key string, value, original any, tp string) {
	if key == "" {
		return
	}

	a.set.AddOriginalAttribute(key, value, original, tp)

	keys := strings.Split(key, ".")
	k := cedar.String(keys[0])
	v := a.set.GetAttributeValue(keys[0])

	a.cedarSet[k] = a.anyToValue(v)
}

// GetAttribute implements the AttributeSet interface.
func (a *attributes) GetAttribute(key string) models.Attribute {
	return a.set.GetAttribute(key)
}

// GetAttributeValue implements the AttributeSet interface.
func (a *attributes) GetAttributeValue(key string) any {
	return a.set.GetAttributeValue(key)
}

// RemoveAttribute implements the AttributeSet interface.
func (a *attributes) RemoveAttribute(key string) {
	a.mutex.Lock()
	a.set.RemoveAttribute(key)
	delete(a.cedarSet, cedar.String(key))
	a.mutex.Unlock()
}

// IterateAttributes implements the AttributeSet interface.
func (a *attributes) IterateAttributes(f models.AttributeIterator) {
	a.set.IterateAttributes(f)
}

// MergeAttributes implements the AttributeSet interface.
func (a *attributes) MergeAttributes(in ...models.AttributeSet) {
	for i := range in {
		a.mutex.Lock()
		in[i].IterateAttributes(func(attr models.Attribute) {
			a.addOriginalAttribute(attr.Key(), attr.Value(), attr.Original(), attr.Type())
		})
		a.mutex.Unlock()
	}
}

// MarshalJSON implements the json.Marshaller interface.
func (a *attributes) MarshalJSON() ([]byte, error) {
	return a.set.MarshalJSON()
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

	case models.AttributeSet:
		s := make(cedar.RecordMap, 0)
		t.IterateAttributes(func(attr models.Attribute) {
			s[cedar.String(attr.Key())] = a.anyToValue(attr.Value())
		})
		return cedar.NewRecord(s)

	default:
		a.logger.Warn("unsupported attribute conversion to Cedar", "type", fmt.Sprintf("%T", in))
		return nil
	}
}

type attributes struct {
	logger   *slog.Logger
	set      models.AttributeSet
	cedarSet cedar.RecordMap
	mutex    sync.RWMutex
}
