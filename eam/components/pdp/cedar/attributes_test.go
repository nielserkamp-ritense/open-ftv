package cedar

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewAttributeBuilder(t *testing.T) {
	t.Run("new attribute builder", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		got := NewAttributeBuilder(logger)
		require.NotNil(t, got)

		a1 := models.NewAttribute("hello", "world")
		a2 := models.NewAttribute("int", 123)
		a3 := models.NewAttribute("float", 123.456)
		a4 := models.NewAttribute("bool", true)

		aa := NewAttributeSet(logger, a3, a4)
		require.NotNil(t, aa)

		got2 := got(a1, aa, a2)
		require.NotNil(t, got2)

		got3, ok := got2.(*attributes)
		require.True(t, ok)
		require.NotNil(t, got3)

		assert.Equal(t, logger, got3.logger)
		assert.Equal(t, 4, len(got3.cedarSet))
	})
}

func TestNewAttributeSet(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	a1 := models.NewAttribute("hello", "world")
	a2 := models.NewAttribute("float", 123.456)
	a3 := models.NewAttribute("bool", false)
	a4 := models.NewAttribute("int", 456)
	a5 := models.NewAttribute("int64", int64(987654321))
	a6 := models.NewAttribute("time", now)
	a7 := models.NewAttribute("duration", time.Second)

	f1, _ := cedar.NewDecimalFromFloat(999.0)

	aa1 := NewAttributeSet(nil, a1, a3)
	aa2 := models.NewAttributeSet(nil, a3, a5, a7)
	aa3 := cedar.RecordMap{"woo": cedar.String("hoo"), "key": f1}

	testCases := []struct {
		name      string
		in        []any
		wantCount int
		want      map[string]any
	}{
		{
			name:      "empty",
			wantCount: 0,
			want:      map[string]any{},
		},
		{
			name:      "1 attribute",
			in:        []any{a1},
			wantCount: 1,
			want:      map[string]any{"hello": "world"},
		},
		{
			name:      "few attributes",
			in:        []any{a1, a2, a3},
			wantCount: 3,
			want: map[string]any{
				"hello": "world",
				"float": 123.456,
				"bool":  false,
			},
		},
		{
			name:      "many attributes",
			in:        []any{a1, a2, a3, a4, a5, a6, a7},
			wantCount: 7,
			want: map[string]any{
				"hello":    "world",
				"float":    123.456,
				"bool":     false,
				"int":      456,
				"int64":    int64(987654321),
				"time":     now,
				"duration": time.Second,
			},
		},
		{
			name:      "1 set",
			in:        []any{aa1},
			wantCount: 2,
			want: map[string]any{
				"hello": "world",
				"bool":  false,
			},
		},
		{
			name:      "few sets",
			in:        []any{aa2, aa1},
			wantCount: 4,
			want: map[string]any{
				"hello":    "world",
				"bool":     false,
				"int64":    int64(987654321),
				"duration": time.Second,
			},
		},
		{
			name:      "mixed",
			in:        []any{a3, aa2, a1, aa3, a2, aa1},
			wantCount: 7,
			want: map[string]any{
				"hello":    "world",
				"float":    123.456,
				"bool":     false,
				"int64":    int64(987654321),
				"duration": time.Second,
				"woo":      "hoo",
				"key":      999.0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			got := NewAttributeSet(logger, tc.in...)
			require.NotNil(t, got)

			got2, ok := got.(*attributes)
			require.True(t, ok)
			require.NotNil(t, got2)

			assert.Equal(t, tc.wantCount, len(got2.cedarSet))

			for key := range tc.want {
				got3 := got.GetAttributeValue(key)
				assert.Equal(t, tc.want[key], got3)
			}

			got.IterateAttributes(func(attr models.Attribute) {
				assert.Equal(t, tc.want[attr.Key()], attr.Value())
			})
		})
	}
}

func TestAttributes_AddAttribute(t *testing.T) {
	testCases := []struct {
		name      string
		add       map[string]any
		wantCount int
		want      map[string]any
	}{
		{
			name:      "none",
			wantCount: 0,
			want:      map[string]any{},
		},
		{
			name:      "one",
			add:       map[string]any{"hello": "world"},
			wantCount: 1,
			want:      map[string]any{"hello": "world"},
		},
		{
			name:      "few",
			add:       map[string]any{"hello": "world", "int": 9, "bool": false},
			wantCount: 3,
			want:      map[string]any{"hello": "world", "bool": false, "int": 9},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			got := NewAttributeSet(logger)
			require.NotNil(t, got)

			for key := range tc.add {
				got.AddAttribute(key, tc.add[key])
			}

			got2, ok := got.(*attributes)
			require.True(t, ok)
			require.NotNil(t, got2)

			assert.Equal(t, tc.wantCount, len(got2.cedarSet))

			for key := range tc.want {
				got3 := got.GetAttributeValue(key)
				assert.Equal(t, tc.want[key], got3)
			}

			got.IterateAttributes(func(attr models.Attribute) {
				assert.Equal(t, tc.want[attr.Key()], attr.Value())
			})
		})
	}
}

func TestAttributes_RemoveAttribute(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	a1 := models.NewAttribute("hello", "world")
	a2 := models.NewAttribute("float", 123.456)
	a3 := models.NewAttribute("bool", false)
	a4 := models.NewAttribute("int", 456)
	a5 := models.NewAttribute("int64", int64(987654321))
	a6 := models.NewAttribute("time", now)
	a7 := models.NewAttribute("duration", time.Second)

	testCases := []struct {
		name      string
		remove    []string
		wantCount int
		want      map[string]any
	}{
		{
			name:      "none",
			wantCount: 7,
			want: map[string]any{
				"hello":    "world",
				"float":    123.456,
				"bool":     false,
				"int":      456,
				"int64":    int64(987654321),
				"time":     now,
				"duration": time.Second,
			},
		},
		{
			name:      "one",
			remove:    []string{"int64"},
			wantCount: 6,
			want: map[string]any{
				"hello":    "world",
				"float":    123.456,
				"bool":     false,
				"int":      456,
				"time":     now,
				"duration": time.Second,
			},
		},
		{
			name:      "few",
			remove:    []string{"int64", "time", "float"},
			wantCount: 4,
			want: map[string]any{
				"hello":    "world",
				"bool":     false,
				"int":      456,
				"duration": time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			got := NewAttributeSet(logger, a1, a2, a3, a4, a5, a6, a7)
			require.NotNil(t, got)

			for i := range tc.remove {
				got.RemoveAttribute(tc.remove[i])
			}

			got2, ok := got.(*attributes)
			require.True(t, ok)
			require.NotNil(t, got2)

			assert.Equal(t, tc.wantCount, len(got2.cedarSet))

			for key := range tc.want {
				got3 := got.GetAttributeValue(key)
				assert.Equal(t, tc.want[key], got3)
			}

			got.IterateAttributes(func(attr models.Attribute) {
				assert.Equal(t, tc.want[attr.Key()], attr.Value())
			})
		})
	}
}

func TestValueToAny(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	f1, _ := cedar.NewDecimalFromFloat(456.789)
	f2, _ := cedar.NewDecimalFromFloat(123.456)

	testCases := []struct {
		name    string
		in      cedar.Value
		want    any
		wantLog int
	}{
		{
			name:    "invalid",
			in:      cedar.EntityUID{Type: "x", ID: "y"},
			want:    `x::"y"`,
			wantLog: 1,
		},
		{
			name: "nil",
		},
		{
			name: "empty string",
			in:   cedar.String(""),
			want: "",
		},
		{
			name: "string",
			in:   cedar.String("world magic"),
			want: "world magic",
		},
		{
			name: "true",
			in:   cedar.Boolean(true),
			want: true,
		},
		{
			name: "false",
			in:   cedar.Boolean(false),
			want: false,
		},
		{
			name: "long",
			in:   cedar.Long(123456),
			want: int64(123456),
		},
		{
			name: "decimal",
			in:   f1,
			want: 456.789,
		},
		{
			name: "time",
			in:   cedar.NewDatetime(now),
			want: now,
		},
		{
			name: "duration",
			in:   cedar.NewDuration(15 * time.Millisecond),
			want: 15 * time.Millisecond,
		},
		{
			name: "empty set",
			in:   cedar.NewSet(),
			want: []any{},
		},
		{
			name: "set",
			in:   cedar.NewSet(cedar.String("yo"), cedar.Boolean(true)),
			want: []any{"yo", true},
		},
		{
			name: "empty map",
			in:   cedar.NewRecord(cedar.RecordMap{}),
			want: map[string]any{},
		},
		{
			name: "map",
			in:   cedar.NewRecord(cedar.RecordMap{"do": cedar.String("it"), "float": f2}),
			want: map[string]any{"do": "it", "float": 123.456},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			got := NewAttributeSet(logger)
			require.NotNil(t, got)

			got2, ok := got.(*attributes)
			require.True(t, ok)
			require.NotNil(t, got2)

			got3 := got2.valueToAny(tc.in)
			assert.Equal(t, tc.wantLog, h.Count())

			s1, ok1 := tc.want.([]any)
			s2, ok2 := got3.([]any)
			if ok1 && ok2 {
				for i := range s1 {
					var found bool
					for j := range s2 {
						if s1[i] == s2[j] {
							found = true
							break
						}
					}
					assert.Truef(t, found, fmt.Sprintf("slices mismatched; %v != %v", s1, s2))
				}
			} else {
				assert.EqualValues(t, tc.want, got3)
			}
		})
	}
}
