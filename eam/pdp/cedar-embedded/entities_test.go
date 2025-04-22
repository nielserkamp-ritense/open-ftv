package cedar_embedded

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	types2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewWrappedEntity(t *testing.T) {
	t.Parallel()

	t.Run("new wrapped entity", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		ce := &cedar.Entity{
			UID:        types.EntityUID{Type: "User", ID: "donaldduck"},
			Attributes: types.NewRecord(cedar.RecordMap{"x": cedar.String("y")}),
			Parents: cedar.NewEntityUIDSet(
				cedar.EntityUID{Type: "User", ID: "marieduck"},
				cedar.EntityUID{Type: "User", ID: "jimmyduck"},
			),
		}

		w := NewWrappedEntity(ce, logger)
		require.NotNil(t, w)

		assert.Equal(t, "User::donaldduck", w.UID())
		assert.Equal(t, "User", w.Type())
		assert.Equal(t, "donaldduck", w.ID())

		attr := w.Attributes()
		require.NotNil(t, attr)

		got, ok := attr.GetAttributeValue("x").(string)
		require.True(t, ok)
		assert.Equal(t, "y", got)

		parents := w.Parents()
		slices.Sort(parents)
		assert.EqualValues(t, []string{"User::jimmyduck", "User::marieduck"}, parents)
	})
}

func TestNewEntitySet(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	e1 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x1"}}
	e2 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x2"}}
	e3 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x3"}}
	e4 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x4"}}
	dup2 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x2"}}

	w1 := NewWrappedEntity(e1, logger)
	w2 := NewWrappedEntity(e2, logger)
	w3 := NewWrappedEntity(e3, logger)
	w4 := NewWrappedEntity(e4, logger)
	dupW2 := NewWrappedEntity(dup2, logger)

	testCases := []struct {
		name string
		in   []any
		want map[string]types2.Entity
	}{
		{
			name: "empty",
		},
		{
			name: "one set",
			in: []any{
				&entities{set: cedar.EntityMap{e1.UID: *e1}},
			},
			want: map[string]types2.Entity{w1.UID(): w1},
		},
		{
			name: "mixed input - no dupes",
			in: []any{
				&entities{set: cedar.EntityMap{e3.UID: *e3, e1.UID: *e1}},
				w2,
				12345,
			},
			want: map[string]types2.Entity{w1.UID(): w1, w2.UID(): w2, w3.UID(): w3},
		},
		{
			name: "mixed input - 1 dupe",
			in: []any{
				w2,
				&entities{set: cedar.EntityMap{e4.UID: *e4, e1.UID: *e1}},
				w3,
				12345,
				dupW2,
				nil,
			},
			want: map[string]types2.Entity{w1.UID(): w1, dupW2.UID(): dupW2, w3.UID(): w3, w4.UID(): w4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewEntitySet(logger, tc.in...)
			require.NotNil(t, got)

			got2, ok := got.(*entities)
			require.True(t, ok)
			require.NotNil(t, got2)

			for k, v := range tc.want {
				v2 := got.GetEntity(k)
				assert.EqualValues(t, v, v2)
			}

			got.IterateEntities(func(entity types2.Entity) {
				assert.EqualValues(t, tc.want[entity.UID()], entity)
			})
		})
	}
}

func TestEntities_AddEntity(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	e1 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x1"}}
	e2 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x2"}}
	e3 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x3"}}
	e4 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x4"}}
	dup2 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x2"}}

	w1 := NewWrappedEntity(e1, logger)
	w2 := NewWrappedEntity(e2, logger)
	w3 := NewWrappedEntity(e3, logger)
	w4 := NewWrappedEntity(e4, logger)
	dupW2 := NewWrappedEntity(dup2, logger)

	testCases := []struct {
		name string
		in   types2.EntitySet
		add  types2.Entity
		want map[string]types2.Entity
	}{
		{
			name: "empty",
			in:   NewEntitySet(logger),
			add:  w3,
			want: map[string]types2.Entity{w3.UID(): w3},
		},
		{
			name: "new key",
			in:   NewEntitySet(logger, w1, w4),
			add:  w2,
			want: map[string]types2.Entity{w1.UID(): w1, w2.UID(): w2, w4.UID(): w4},
		},
		{
			name: "duplicate key",
			in:   NewEntitySet(logger, &entities{set: cedar.EntityMap{e2.UID: *e2, e3.UID: *e3}}),
			add:  dupW2,
			want: map[string]types2.Entity{dupW2.UID(): dupW2, w3.UID(): w3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := tc.in
			e.AddEntity(tc.add)

			for k, v := range tc.want {
				v2 := e.GetEntity(k)
				assert.EqualValues(t, v, v2)
			}

			e.IterateEntities(func(entity types2.Entity) {
				assert.EqualValues(t, tc.want[entity.UID()], entity)
			})
		})
	}
}

func TestEntities_RemoveEntity(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	e1 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x1"}}
	e2 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x2"}}
	e3 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x3"}}
	e4 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x4"}}

	w1 := NewWrappedEntity(e1, logger)
	w2 := NewWrappedEntity(e2, logger)
	w3 := NewWrappedEntity(e3, logger)
	w4 := NewWrappedEntity(e4, logger)

	testCases := []struct {
		name string
		in   types2.EntitySet
		key  string
		want map[string]types2.Entity
	}{
		{
			name: "empty",
			in:   NewEntitySet(logger),
			key:  "hello::x1",
			want: map[string]types2.Entity{},
		},
		{
			name: "miss",
			in:   NewEntitySet(logger, &entities{set: cedar.EntityMap{e1.UID: *e1, e4.UID: *e4}}),
			key:  "entity::x2",
			want: map[string]types2.Entity{w1.UID(): w1, w4.UID(): w4},
		},
		{
			name: "hit",
			in:   NewEntitySet(logger, w2, w3),
			key:  "entity::x2",
			want: map[string]types2.Entity{w3.UID(): w3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := tc.in
			e.RemoveEntity(tc.key)

			for k, v := range tc.want {
				v2 := e.GetEntity(k)
				assert.EqualValues(t, v, v2)
			}

			e.IterateEntities(func(entity types2.Entity) {
				assert.EqualValues(t, tc.want[entity.UID()], entity)
			})
		})
	}
}

func TestEntities_MergeEntities(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	e1 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x1"}}
	e2 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x2"}}
	e3 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x3"}}
	e4 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x4"}}
	e5 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x5"}}
	e6 := &cedar.Entity{UID: cedar.EntityUID{Type: "entity", ID: "x6"}}

	w1 := NewWrappedEntity(e1, logger)
	w2 := NewWrappedEntity(e2, logger)
	w3 := NewWrappedEntity(e3, logger)
	w4 := NewWrappedEntity(e4, logger)
	w5 := NewWrappedEntity(e5, logger)
	w6 := NewWrappedEntity(e6, logger)

	testCases := []struct {
		name  string
		in    types2.EntitySet
		merge []types2.EntitySet
		want  map[string]types2.Entity
	}{
		{
			name: "both empty",
			in:   NewEntitySet(logger),
			want: make(map[string]types2.Entity),
		},
		{
			name:  "add one set",
			in:    NewEntitySet(logger, w1, w4),
			merge: []types2.EntitySet{NewEntitySet(logger, w2, w3)},
			want:  map[string]types2.Entity{w1.UID(): w1, w2.UID(): w2, w3.UID(): w3, w4.UID(): w4},
		},
		{
			name:  "add few sets",
			in:    NewEntitySet(logger, w1, w4),
			merge: []types2.EntitySet{NewEntitySet(logger, w5, w3), NewEntitySet(logger, w6, w4)},
			want:  map[string]types2.Entity{w1.UID(): w1, w3.UID(): w3, w4.UID(): w4, w5.UID(): w5, w6.UID(): w6},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := tc.in
			e.MergeEntities(tc.merge...)

			for k, v := range tc.want {
				v2 := e.GetEntity(k)
				assert.EqualValues(t, v, v2)
			}

			e.IterateEntities(func(entity types2.Entity) {
				assert.EqualValues(t, tc.want[entity.UID()], entity)
			})
		})
	}
}
