package pip

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestSourceRefs(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	p := New(context.Background(), logger)

	refs, ok := p.(SourceReferencer)
	require.True(t, ok, "PIP must implement SourceReferencer")

	// unknown key.
	_, found := refs.SourceRef("nope")
	assert.False(t, found)

	// record and read back.
	refs.RecordSourceRef(SourceRef{
		Kind:     "attribute",
		Key:      "leeftijd",
		TraceID:  "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:   "00f067aa0ba902b7",
		WARCFile: "pip-20260708.warc",
		Version:  "2.0.0",
		Sequence: "7",
	})

	ref, found := refs.SourceRef("leeftijd")
	require.True(t, found)
	assert.Equal(t, "attribute", ref.Kind)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", ref.TraceID)
	assert.Equal(t, "00f067aa0ba902b7", ref.SpanID)
	assert.Equal(t, "pip-20260708.warc", ref.WARCFile)
	assert.Equal(t, "2.0.0", ref.Version)
	assert.Equal(t, "7", ref.Sequence)
	assert.False(t, ref.Time.IsZero(), "Time must be stamped on record")

	// replace keeps the latest version.
	refs.RecordSourceRef(SourceRef{Kind: "attribute", Key: "leeftijd", Version: "2.0.1"})
	ref, found = refs.SourceRef("leeftijd")
	require.True(t, found)
	assert.Equal(t, "2.0.1", ref.Version)

	// list contains the reference.
	list := refs.ListSourceRefs()
	require.Len(t, list, 1)
	assert.Equal(t, "leeftijd", list[0].Key)
}

type testSink struct {
	mutex  sync.Mutex
	events []models.EventType
	keys   []string
}

func (s *testSink) Handle(t models.EventType, key string) {
	s.mutex.Lock()
	s.events = append(s.events, t)
	s.keys = append(s.keys, key)
	s.mutex.Unlock()
}

func TestWithEventSink(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	sink := &testSink{}
	p := New(context.Background(), logger, WithEventSink(sink))

	p.AddAttribute("a", 1)                                                  // added.
	p.AddAttribute("a", 2)                                                  // replaced.
	p.RemoveAttribute("a")                                                  // removed.
	p.AddEntity(models.NewEntity("service", "x", models.NewAttributeSet())) // added.
	p.RemoveEntity(models.EntityUID("service", "x"))                        // removed.

	sink.mutex.Lock()
	defer sink.mutex.Unlock()

	require.Equal(t, []models.EventType{
		models.AttributeAdded,
		models.AttributeReplaced,
		models.AttributeRemoved,
		models.EntityAdded,
		models.EntityRemoved,
	}, sink.events)
	assert.Equal(t, "a", sink.keys[0])
	assert.Equal(t, models.EntityUID("service", "x"), sink.keys[3])
}
