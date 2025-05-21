package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestRecord_FieldString(t *testing.T) {
	t.Parallel()

	r1 := &Row{
		Data: map[string]any{
			"foo":       "bar",
			"int":       123,
			"float":     567.89,
			"bool":      true,
			"timestamp": time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	testCases := []struct {
		name string
		rec  *Row
		id   string
		want string
	}{
		{name: "empty", rec: r1},
		{name: "not found", rec: r1, id: "yo"},
		{name: "record empty", rec: &Row{Data: map[string]any{}}, id: "foo"},
		{name: "string", rec: r1, id: "foo", want: "bar"},
		{name: "int", rec: r1, id: "int", want: "123"},
		{name: "float", rec: r1, id: "float", want: "567.89"},
		{name: "bool", rec: r1, id: "bool", want: "true"},
		{name: "timestamp", rec: r1, id: "timestamp", want: "2020-02-01 00:00:00 +0000 UTC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.rec.FieldString(tc.id)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRecordFromData(t *testing.T) {
	t.Parallel()

	t1 := &schema.Object{
		Parent:      schema.Parent{ID: "t1"},
		Description: "t1",
		Fields: []*schema.Field{
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f1"}},
				Type:   types.StringType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f2"}},
				Type:   types.IntegerType,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f3"}},
				Type:   types.DateType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f4"}},
				Type:   types.EmailType,
				IsPII:  true,
			},
		},
	}

	testCases := []struct {
		name string
		t    *schema.Object
		data map[string]any
		want *Row
	}{
		{
			name: "all empty",
			t:    t1,
			data: map[string]any{},
			want: &Row{Data: map[string]any{}},
		},
		{
			name: "unknown fields",
			t:    t1,
			data: map[string]any{"hello": "world"},
			want: &Row{Data: map[string]any{}},
		},
		{
			name: "f1",
			t:    t1,
			data: map[string]any{"f1": "hello world"},
			want: &Row{Data: map[string]any{"f1": "hello world"}},
		},
		{
			name: "all",
			t:    t1,
			data: map[string]any{
				"f1": "hello world",
				"f2": 123,
				"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				"f4": "joep@tv.nl",
			},
			want: &Row{
				Data: map[string]any{
					"f1": "hello world",
					"f2": int64(123),
					"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
					"f4": "joep@tv.nl",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RowFromData(tc.t, tc.data)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRecordFromCSV(t *testing.T) {
	t.Parallel()

	t1 := &schema.Object{
		Parent:      schema.Parent{ID: "t1"},
		Description: "t1",
		Fields: []*schema.Field{
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f1"}},
				Type:   types.StringType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f2"}},
				Type:   types.IntegerType,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f3"}},
				Type:   types.DateType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f4"}},
				Type:   types.EmailType,
				IsPII:  true,
			},
		},
	}

	testCases := []struct {
		name    string
		t       *schema.Object
		headers []string
		data    []string
		want    *Row
	}{
		{
			name:    "all empty",
			t:       t1,
			headers: []string{},
			data:    []string{},
			want:    &Row{Data: map[string]any{}},
		},
		{
			name:    "unknown fields",
			t:       t1,
			headers: []string{"hello"},
			data:    []string{"world"},
			want:    &Row{Data: map[string]any{}},
		},
		{
			name:    "f1",
			t:       t1,
			headers: []string{"f1"},
			data:    []string{"hello world"},
			want:    &Row{Data: map[string]any{"f1": "hello world"}},
		},
		{
			name:    "all",
			t:       t1,
			headers: []string{"f1", "f2", "f3", "f4"},
			data:    []string{"hello world", "123", "2024-06-10", "joep@tv.nl"},
			want: &Row{
				Data: map[string]any{
					"f1": "hello world",
					"f2": int64(123),
					"f3": time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
					"f4": "joep@tv.nl",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RowFromCSV(tc.t, tc.headers, tc.data)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRecord_MatchFilter(t *testing.T) {
	t.Parallel()

	var r1 = &Row{Data: map[string]any{
		"f1": "hello world",
		"f2": 123,
		"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		"f4": true,
		"f5": 98765.4321,
	}}

	testCases := []struct {
		name   string
		record *Row
		filter map[string]any
		want   bool
	}{
		{
			name:   "empty filter",
			record: r1,
			want:   true,
		},
		{
			name:   "matched string",
			record: r1,
			filter: map[string]any{"f1": "hello world"},
			want:   true,
		},
		{
			name:   "unmatched string",
			record: r1,
			filter: map[string]any{"f1": "hello worlds"},
		},
		{
			name:   "matched regex",
			record: r1,
			filter: map[string]any{"f1": regexp.MustCompile(".*ello.*")},
			want:   true,
		},
		{
			name:   "unmatched regex",
			record: r1,
			filter: map[string]any{"f1": regexp.MustCompile(".*elo.*")},
		},
		{
			name:   "matched int",
			record: r1,
			filter: map[string]any{"f2": "123"},
			want:   true,
		},
		{
			name:   "unmatched int",
			record: r1,
			filter: map[string]any{"f2": 124},
		},
		{
			name:   "matched date",
			record: r1,
			filter: map[string]any{"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
			want:   true,
		},
		{
			name:   "unmatched date",
			record: r1,
			filter: map[string]any{"f3": time.Date(2024, 6, 1, 0, 0, 0, 1, time.UTC)},
		},
		{
			name:   "matched float",
			record: r1,
			filter: map[string]any{"f5": 98765.4321},
			want:   true,
		},
		{
			name:   "unmatched float",
			record: r1,
			filter: map[string]any{"f5": 98765.432},
		},
		{
			name:   "all matched",
			record: r1,
			filter: map[string]any{
				"f5": 98765.4321,
				"f2": "123",
				"f1": regexp.MustCompile(".*ello.*"),
				"f4": true,
				"f3": regexp.MustCompile("2024-06*"),
			},
			want: true,
		},
		{
			name:   "one unmatched",
			record: r1,
			filter: map[string]any{
				"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				"f2": "123",
				"f1": regexp.MustCompile(".*ello.*"),
				"f4": "1", // this is not "true"
				"f5": 98765.4321},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.record.MatchFilter(tc.filter)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
