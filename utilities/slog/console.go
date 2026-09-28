package slog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// ConsoleHandler is safe to use with concurrent go routines.
type ConsoleHandler struct {
	w      io.Writer
	level  slog.Leveler
	source bool
	prefix string
	// attrs holds the handler attributes already written out, so the group prefix they were added under
	// is applied exactly once.
	attrs []byte
	mutex *sync.Mutex
}

// NewConsoleHandler instantiates a new compact log handler writing to w. A nil opts logs at info level without
// the source.
func NewConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *ConsoleHandler {
	c := &ConsoleHandler{w: w, level: slog.LevelInfo, mutex: &sync.Mutex{}}
	if opts != nil {
		c.source = opts.AddSource
		if opts.Level != nil {
			c.level = opts.Level
		}
	}

	return c
}

// Enabled implements the slog.Handler interface.
func (c *ConsoleHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= c.level.Level()
}

// Handle implements the slog.Handler interface.
func (c *ConsoleHandler) Handle(_ context.Context, rec slog.Record) error { //nolint:gocritic // slog.Handler dictates passing the record by value.
	buf := bytes.Buffer{}
	buf.WriteString(rec.Level.String())
	buf.WriteByte(' ')
	buf.WriteString(rec.Message)
	buf.Write(c.attrs)

	rec.Attrs(func(attr slog.Attr) bool {
		writeAttr(&buf, c.prefix, attr)
		return true
	})

	if c.source && rec.PC != 0 {
		writeSource(&buf, rec.PC)
	}

	buf.WriteByte('\n')

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.w.Write(buf.Bytes())

	return err
}

// WithAttrs implements the slog.Handler interface.
func (c *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return c
	}

	clone := c.clone()

	// clone.attrs has no spare capacity, so writing to it never touches the parent's backing array.
	buf := bytes.NewBuffer(clone.attrs)
	for i := range attrs {
		writeAttr(buf, c.prefix, attrs[i])
	}

	clone.attrs = buf.Bytes()

	return clone
}

// WithGroup implements the slog.Handler interface.
func (c *ConsoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return c
	}

	clone := c.clone()
	clone.prefix += name + "."

	return clone
}

func (c *ConsoleHandler) clone() *ConsoleHandler {
	return &ConsoleHandler{
		w:      c.w,
		level:  c.level,
		source: c.source,
		prefix: c.prefix,
		attrs:  c.attrs[:len(c.attrs):len(c.attrs)],
		mutex:  c.mutex,
	}
}

// prefixAttr resolves the value of an attribute and prepends the given group prefix to its key.
func prefixAttr(prefix string, attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	attr.Key = prefix + attr.Key

	return attr
}

// writeAttr appends an attribute to the buffer as ` key=value`, recursing into groups.
func writeAttr(buf *bytes.Buffer, prefix string, attr slog.Attr) {
	// Empty attributes are ignored, in line with the standard slog handlers.
	if attr.Equal(slog.Attr{}) {
		return
	}

	attr = prefixAttr(prefix, attr)

	if attr.Value.Kind() == slog.KindGroup {
		group := attr.Value.Group()

		// A group without a key is inlined, a group without attributes is skipped.
		groupPrefix := prefix
		if attr.Key != prefix {
			groupPrefix = attr.Key + "."
		}

		for i := range group {
			writeAttr(buf, groupPrefix, group[i])
		}

		return
	}

	buf.WriteByte(' ')
	buf.WriteString(attr.Key)
	buf.WriteByte('=')
	buf.WriteString(quoteValue(attr.Value.String()))
}

// writeSource appends the call site as ` source=dir/file.go:line`: the parent directory keeps the many main.go and
// config.go files apart without the full path.
func writeSource(buf *bytes.Buffer, pc uintptr) {
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	if frame.File == "" {
		return
	}

	buf.WriteString(" source=")
	buf.WriteString(filepath.Base(filepath.Dir(frame.File)))
	buf.WriteByte('/')
	buf.WriteString(filepath.Base(frame.File))
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(frame.Line))
}

// quoteValue quotes a value if it would otherwise be hard to tell apart from the rest of the line, or would
// split the record over several lines.
func quoteValue(value string) string {
	if value == "" || strings.ContainsFunc(value, needsQuote) {
		return strconv.Quote(value)
	}

	return value
}

func needsQuote(r rune) bool {
	return r == '"' || r == '=' || unicode.IsSpace(r) || !unicode.IsPrint(r)
}
