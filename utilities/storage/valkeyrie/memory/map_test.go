package memory

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/kvtools/valkeyrie/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new", func(t *testing.T) {
		s := New()
		require.NotNil(t, s)

		defer s.Close()

		s2, ok := s.(*kv)
		require.True(t, ok)
		require.NotNil(t, s2)
		require.NotNil(t, s2.kv)
	})
}

func TestPutGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		add  map[string]any
	}{
		{name: "single simple value", add: singleSimple},
		{name: "single object", add: singleObject},
		{name: "few simple values", add: fewSimple},
		{name: "few objects", add: fewObjects},
		{name: "mixed", add: mixed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			for k, v := range tc.add {
				pair, err := s.Get(nil, k, nil)
				require.NoError(t, err)
				assert.NotNil(t, pair)
				assert.Equal(t, k, pair.Key)

				d, _ := json.Marshal(v)
				assert.Equal(t, d, pair.Value)
			}
		})
	}
}

func TestGetFail(t *testing.T) {
	t.Parallel()

	t.Run("get fail", func(t *testing.T) {
		s := New()
		require.NotNil(t, s)

		defer s.Close()

		s2, ok := s.(*kv)
		require.True(t, ok)
		require.NotNil(t, s2)
		require.NotNil(t, s2.kv)

		pair, err := s.Get(nil, "a", nil)
		require.Error(t, err)
		require.Nil(t, pair)
	})
}

func TestPutDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		add  map[string]any
	}{
		{name: "single simple value", add: singleSimple},
		{name: "single object", add: singleObject},
		{name: "mixed", add: mixed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			s2, ok := s.(*kv)
			require.True(t, ok)
			require.NotNil(t, s2)
			assert.Equal(t, len(tc.add), len(s2.kv))

			for k := range tc.add {
				err := s.Delete(nil, k)
				require.NoError(t, err)
			}

			assert.Zero(t, len(s2.kv))

			for k := range tc.add {
				pair, err := s.Get(nil, k, nil)
				require.Error(t, err)
				assert.Nil(t, pair)
			}
		})
	}
}

func TestPutExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		add  map[string]any
	}{
		{name: "single simple value", add: singleSimple},
		{name: "single object", add: singleObject},
		{name: "mixed", add: mixed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			s2, ok := s.(*kv)
			require.True(t, ok)
			require.NotNil(t, s2)
			assert.Equal(t, len(tc.add), len(s2.kv))

			for k := range tc.add {
				found, err := s.Exists(nil, k, nil)
				require.NoError(t, err)
				require.True(t, found)
			}
		})
	}
}

func TestExistsFail(t *testing.T) {
	t.Run("exists fail", func(t *testing.T) {
		s := New()
		require.NotNil(t, s)

		s2, ok := s.(*kv)
		require.True(t, ok)
		require.NotNil(t, s2)
		require.NotNil(t, s2.kv)

		found, err := s.Exists(nil, "a", nil)
		require.Error(t, err)
		require.False(t, found)
	})
}

func TestUnsupported(t *testing.T) {
	t.Parallel()

	t.Run("unsupported", func(t *testing.T) {
		s := New()
		require.NotNil(t, s)

		defer s.Close()

		ch, err := s.Watch(nil, "a", nil)
		require.Error(t, err)
		require.Nil(t, ch)

		ch2, err2 := s.WatchTree(nil, "a", nil)
		require.Error(t, err2)
		require.Nil(t, ch2)

		lock, err3 := s.NewLock(nil, "a", nil)
		require.Error(t, err3)
		require.Nil(t, lock)
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		add       map[string]any
		directory string
		wantCount int
	}{
		{name: "single simple value - no directory", add: singleSimple, wantCount: 1},
		{name: "single simple value - bad directory", add: singleSimple, directory: "b"},
		{name: "single simple value - good directory", add: singleSimple, directory: "a", wantCount: 1},
		{name: "single object", add: singleObject, directory: "a", wantCount: 1},
		{name: "mixed", add: mixed, directory: "c", wantCount: 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			list, err := s.List(nil, tc.directory, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCount, len(list))
		})
	}
}

func TestDeleteTree(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		add       map[string]any
		directory string
		wantCount int
	}{
		{name: "single simple value - no directory", add: singleSimple},
		{name: "single simple value - bad directory", add: singleSimple, directory: "b", wantCount: 1},
		{name: "single simple value - good directory", add: singleSimple, directory: "a"},
		{name: "single object", add: singleObject, directory: "b", wantCount: 1},
		{name: "mixed", add: mixed, directory: "c", wantCount: 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			err := s.DeleteTree(nil, tc.directory)
			require.NoError(t, err)

			list, err2 := s.List(nil, "", nil)
			require.NoError(t, err2)
			assert.Equal(t, tc.wantCount, len(list))
		})
	}
}

func TestAtomicPut(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		add       map[string]any
		prevKey   string
		prevValue any
		newKey    string
		newValue  any
		wantErr   bool
	}{
		{name: "create ok", add: mixed, newKey: "i", newValue: true},
		{name: "create duplicate", add: mixed, newKey: "a/h", newValue: true, wantErr: true},
		{name: "update ok", add: mixed, prevKey: "b/e", prevValue: 1, newKey: "b/e", newValue: true},
		{name: "update key mismatch", add: mixed, prevKey: "b/e", prevValue: 1, newKey: "b/f", newValue: true, wantErr: true},
		{name: "update previous mismatch", add: mixed, prevKey: "b/e", prevValue: 2, newKey: "b/e", newValue: true, wantErr: true},
		{name: "update not found", add: mixed, prevKey: "i", prevValue: 1, newKey: "i", newValue: true, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			var prev *store.KVPair
			if tc.prevKey != "" {
				b, err := json.Marshal(tc.prevValue)
				require.NoError(t, err)

				prev = &store.KVPair{Key: tc.prevKey, Value: b}
			}

			b, err := json.Marshal(tc.newValue)
			require.NoError(t, err)

			found, pair, err2 := s.AtomicPut(nil, tc.newKey, b, prev, nil)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, pair)
				require.False(t, found)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, pair)
				require.True(t, found)
			}
		})
	}

}

func TestAtomicDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		add       map[string]any
		prevKey   string
		prevValue any
		key       string
		wantErr   bool
	}{
		{name: "delete ok", add: mixed, prevKey: "b/e", prevValue: 1, key: "b/e"},
		{name: "delete previous mismatch", add: mixed, prevKey: "b/e", prevValue: 2, key: "b/e", wantErr: true},
		{name: "delete not found", add: mixed, prevKey: "i", prevValue: 1, key: "i", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New()
			require.NotNil(t, s)

			defer s.Close()

			for k, v := range tc.add {
				d, err := json.Marshal(v)
				require.NoError(t, err)

				err = s.Put(nil, k, d, nil)
				require.NoError(t, err)
			}

			var prev *store.KVPair
			if tc.prevKey != "" {
				b, err := json.Marshal(tc.prevValue)
				require.NoError(t, err)

				prev = &store.KVPair{Key: tc.prevKey, Value: b}
			}

			found, err := s.AtomicDelete(nil, tc.key, prev)
			if tc.wantErr {
				require.Error(t, err)
				require.False(t, found)
			} else {
				require.NoError(t, err)
				require.True(t, found)
			}
		})
	}

}

var (
	singleSimple = map[string]any{"a/a": 1}
	singleObject = map[string]any{"a/b": map[string]any{"b": 1, "hello": "world", "bool": true}}
	fewSimple    = map[string]any{"a": 1, "b": 2.2, "c": true, "d": "hello world"}

	fewObjects = map[string]any{
		"a": map[string]any{"b": 1, "hello": "world", "bool": true},
		"b": map[string]any{"b": 2, "hello2": "world", "bool": false},
		"c": map[string]any{"b": 3, "what": "words", "float": 1.2345678},
		"d": map[string]any{"b": 4, "donald": "duck", "map": map[string]any{"a": 1, "bool": true}},
	}

	mixed = map[string]any{
		"a/h": map[string]any{"b": 4, "donald": "duck", "map": map[string]any{"a": 1, "bool": true}},
		"a/c": true,
		"d":   "hello world",
		"b/a": map[string]any{"b": 1, "hello": "world", "bool": true},
		"b/e": 1,
		"c/b": map[string]any{"b": 2, "hello2": "world", "bool": false},
		"c/g": map[string]any{"b": 3, "what": "words", "float": 1.2345678},
		"c/f": 2.2,
	}
)
