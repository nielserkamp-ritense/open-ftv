package csv

import (
	"bytes"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// BytesEncoder represents the interface for encoding field values in CSV format to a bytes buffer.
type BytesEncoder interface {
	WriteString(string)        // write a quoted string.
	WriteSeparator()           // write a comma.
	WriteEOL()                 // replace the last character with a newline character.
	Encode(*schema.Field, any) // encode a field value.
	Bytes() []byte             // retrieve the result buffer.
}

// NewBytesEncoder implements a new CSV encoder using a bytes buffer.
func NewBytesEncoder() BytesEncoder {
	return &encoder{}
}

type encoder struct {
	bytes.Buffer
}

// WriteString implements the CSV Encoder interface.
func (e *encoder) WriteString(v string) {
	e.Buffer.WriteString(strconv.Quote(v))
}

// WriteSeparator implements the CSV Encoder interface.
func (e *encoder) WriteSeparator() {
	e.WriteByte(',')
}

// WriteEOL implements the CSV Encoder interface.
func (e *encoder) WriteEOL() {
	e.Truncate(e.Len() - 1)
	e.WriteByte('\n')
}

// Encode implements the CSV Encoder interface.
func (e *encoder) Encode(def *schema.Field, v any) {
	v = encode(def, v)

	switch t := v.(type) {
	case string:
		e.WriteString(t)
	case int64:
		e.Buffer.WriteString(strconv.FormatInt(t, 10))
	case uint64:
		e.Buffer.WriteString(strconv.FormatUint(t, 10))
	case float64:
		e.Buffer.WriteString(strconv.FormatFloat(t, 'g', -1, 64))
	case bool:
		e.WriteString(encodeBool(t))
	case time.Time:
		e.WriteString(encodeTime(def, t))
	default:
		b, _ := json.Marshal(v)
		e.WriteString(string(b))
	}
}

// Bytes implements the CSV Encoder interface.
func (e *encoder) Bytes() []byte {
	return e.Buffer.Bytes()
}
