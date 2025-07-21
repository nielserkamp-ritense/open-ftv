package pap

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestClearWatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		paths []string
	}{
		{
			name:  "nil",
			paths: nil,
		},
		{
			name:  "one path",
			paths: []string{"../../testdata/policies"},
		},
		{
			name:  "few paths",
			paths: []string{"../../testdata/policies", "../../testdata/policies/opa", "../../testdata/policies/cedar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &PAP{}

			if len(tc.paths) > 0 {
				p.watcher, _ = fsnotify.NewWatcher()

				for i := range tc.paths {
					_ = p.watcher.Add(tc.paths[i])
				}
			}

			p.clearWatcher()
			if len(tc.paths) == 0 {
				assert.Nil(t, p.watcher)
			} else {
				assert.Empty(t, p.watcher.WatchList())
			}
		})
	}
}

func TestProcessDeletes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		files map[string]struct{}
	}{
		{
			name:  "empty",
			files: map[string]struct{}{},
		},
		{
			name:  "one",
			files: map[string]struct{}{"x.txt": {}},
		},
		{
			name:  "few",
			files: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := memory.New()

			p := &PAP{
				deletes: tc.files,
				store:   s,
				persist: NewStore(nil, s, ""),
			}

			p.processDeletes()
			assert.Empty(t, p.deletes)
		})
	}
}

func TestProcessUpdates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		files map[string]struct{}
	}{
		{
			name:  "empty",
			files: map[string]struct{}{},
		},
		{
			name:  "one",
			files: map[string]struct{}{"x.txt": {}},
		},
		{
			name:  "few",
			files: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := memory.New()

			p := &PAP{
				updates: tc.files,
				store:   s,
				persist: NewStore(nil, s, ""),
			}

			p.processUpdates()
			assert.Empty(t, p.updates)
		})
	}
}

func TestPolicyModified(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		create      []string
		write       []string
		rename      []string
		remove      []string
		wantUpdates map[string]struct{}
		wantDeletes map[string]struct{}
	}{
		{
			name:        "none",
			wantUpdates: map[string]struct{}{},
			wantDeletes: map[string]struct{}{},
		},
		{
			name:        "create",
			create:      []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}},
			wantDeletes: map[string]struct{}{},
		},
		{
			name:        "write",
			write:       []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}},
			wantDeletes: map[string]struct{}{},
		},
		{
			name:        "rename",
			rename:      []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}},
			wantDeletes: map[string]struct{}{},
		},
		{
			name:        "remove",
			remove:      []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: map[string]struct{}{},
			wantDeletes: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}},
		},
		{
			name:        "mixed",
			create:      []string{"a.txt", "b.txt"},
			write:       []string{"c.txt", "d.txt", "f.txt"},
			rename:      []string{"f.txt", "e.txt", "c.txt"},
			remove:      []string{"a.txt", "f.txt", "c.txt"},
			wantUpdates: map[string]struct{}{"a.txt": {}, "b.txt": {}, "c.txt": {}, "d.txt": {}, "e.txt": {}, "f.txt": {}},
			wantDeletes: map[string]struct{}{"a.txt": {}, "c.txt": {}, "f.txt": {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := memory.New()

			p := &PAP{
				store:   s,
				persist: NewStore(nil, s, ""),
				updates: map[string]struct{}{},
				deletes: map[string]struct{}{},
			}

			for i := range tc.create {
				p.policyModified(fsnotify.Event{Name: tc.create[i], Op: fsnotify.Create})
			}

			for i := range tc.write {
				p.policyModified(fsnotify.Event{Name: tc.write[i], Op: fsnotify.Write})
			}

			for i := range tc.rename {
				p.policyModified(fsnotify.Event{Name: tc.rename[i], Op: fsnotify.Rename})
			}

			for i := range tc.remove {
				p.policyModified(fsnotify.Event{Name: tc.remove[i], Op: fsnotify.Remove})
			}

			assert.EqualValues(t, tc.wantUpdates, p.updates)
			assert.EqualValues(t, tc.wantDeletes, p.deletes)
		})
	}
}

func TestWatchFiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		create []string
		remove []string
		want   int
	}{
		{
			name: "none",
		},
		{
			name:   "one create",
			create: []string{"a.txt"},
			want:   1,
		},
		{
			name:   "few creates",
			create: []string{"a.txt", "b.txt", "c.txt"},
			want:   3,
		},
		{
			name:   "duplicate create",
			create: []string{"a.txt", "b.txt", "a.txt"},
			want:   2,
		},
		{
			name:   "one delete",
			create: []string{"a.txt", "b.txt", "c.txt"},
			remove: []string{"b.txt"},
			want:   2,
		},
		{
			name:   "few deletes",
			create: []string{"a.txt", "b.txt", "c.txt"},
			remove: []string{"c.txt", "b.txt", "a.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dir := t.TempDir()

			w, err := fsnotify.NewWatcher()
			require.NoError(t, err)
			require.NotNil(t, w)

			err = w.Add(dir)
			require.NoError(t, err)

			s := memory.New()

			p := &PAP{
				ctx:     ctx,
				watcher: w,
				store:   s,
				persist: NewStore(nil, s, ""),
				updates: map[string]struct{}{},
				deletes: map[string]struct{}{},
			}

			wg := &sync.WaitGroup{}
			wg.Add(2)

			go func(wg *sync.WaitGroup) {
				p.watchFiles()
				wg.Done()
			}(wg)

			go func(wg *sync.WaitGroup) {
				for i := range tc.create {
					f, err2 := os.Create(filepath.Join(dir, tc.create[i]))
					require.NoError(t, err2)

					_, err2 = f.Write([]byte("x"))
					require.NoError(t, err2)

					f.Close()
				}

				for i := range tc.remove {
					err2 := os.Remove(filepath.Join(dir, tc.remove[i]))
					require.NoError(t, err2)
				}

				wg.Done()
			}(wg)

			for range 15 {
				time.Sleep(25 * time.Millisecond)
			}

			cancel()

			wg.Wait()

			assert.Nil(t, p.watcher)

			p.mutex.Lock()
			assert.Empty(t, p.updates)
			assert.Empty(t, p.deletes)
			p.mutex.Unlock()
		})
	}
}
