package pip

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestClearEntityWatcher(t *testing.T) {
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
			paths: []string{"../../../testdata/pip"},
		},
		{
			name:  "few paths",
			paths: []string{"../../../testdata/pip", "../../../testdata/pip/opa", "../../../testdata/pip/cedar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &pip{}

			if len(tc.paths) > 0 {
				p.entityWatcher, _ = fsnotify.NewWatcher()

				for i := range tc.paths {
					_ = p.entityWatcher.Add(tc.paths[i])
				}
			}

			p.clearEntityWatcher()
			if len(tc.paths) == 0 {
				assert.Nil(t, p.entityWatcher)
			} else {
				assert.Empty(t, p.entityWatcher.WatchList())
			}
		})
	}
}

func TestEntityDeletes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		files []string
	}{
		{
			name:  "empty",
			files: []string{},
		},
		{
			name:  "one",
			files: []string{"x.txt"},
		},
		{
			name:  "few",
			files: []string{"a.txt", "b.txt", "c.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &pip{entityDeletes: tc.files}
			p.processEntityDeletes()
			assert.Empty(t, p.entityDeletes)
		})
	}
}

func TestEntityUpdates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		files []string
	}{
		{
			name:  "empty",
			files: []string{},
		},
		{
			name:  "one",
			files: []string{"x.txt"},
		},
		{
			name:  "few",
			files: []string{"a.txt", "b.txt", "c.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &pip{entityUpdates: tc.files}
			p.processEntityUpdates()
			assert.Empty(t, p.entityUpdates)
		})
	}
}

func TestEntitiesModified(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		create      []string
		write       []string
		rename      []string
		remove      []string
		wantUpdates []string
		wantDeletes []string
	}{
		{
			name: "none",
		},
		{
			name:        "create",
			create:      []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: []string{"a.txt", "b.txt", "c.txt"},
		},
		{
			name:        "write",
			write:       []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: []string{"a.txt", "b.txt", "c.txt"},
		},
		{
			name:        "rename",
			rename:      []string{"a.txt", "b.txt", "c.txt"},
			wantUpdates: []string{"a.txt", "b.txt", "c.txt"},
		},
		{
			name:        "remove",
			remove:      []string{"a.txt", "b.txt", "c.txt"},
			wantDeletes: []string{"a.txt", "b.txt", "c.txt"},
		},
		{
			name:        "mixed",
			create:      []string{"a.txt", "b.txt"},
			write:       []string{"c.txt", "d.txt", "f.txt"},
			rename:      []string{"f.txt", "e.txt", "c.txt"},
			remove:      []string{"a.txt", "f.txt", "c.txt"},
			wantUpdates: []string{"a.txt", "b.txt", "c.txt", "d.txt", "e.txt", "f.txt"},
			wantDeletes: []string{"a.txt", "c.txt", "f.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &pip{}

			for i := range tc.create {
				p.entitiesModified(fsnotify.Event{Name: tc.create[i], Op: fsnotify.Create})
			}

			for i := range tc.write {
				p.entitiesModified(fsnotify.Event{Name: tc.write[i], Op: fsnotify.Write})
			}

			for i := range tc.rename {
				p.entitiesModified(fsnotify.Event{Name: tc.rename[i], Op: fsnotify.Rename})
			}

			for i := range tc.remove {
				p.entitiesModified(fsnotify.Event{Name: tc.remove[i], Op: fsnotify.Remove})
			}

			assert.EqualValues(t, tc.wantUpdates, p.entityUpdates)
			assert.EqualValues(t, tc.wantDeletes, p.entityDeletes)
		})
	}
}

func TestWatchEntityFiles(t *testing.T) {
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
			want:   3,
		},
		{
			name:   "few deletes",
			create: []string{"a.txt", "b.txt", "c.txt"},
			remove: []string{"c.txt", "b.txt", "a.txt"},
			want:   3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := util.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			dir := t.TempDir()

			w, err := fsnotify.NewWatcher()
			require.NoError(t, err)
			require.NotNil(t, w)

			err = w.Add(dir)
			require.NoError(t, err)

			p := New(ctx, logger).(*pip)
			require.NotNil(t, p)

			p.entityWatcher = w

			wg := &sync.WaitGroup{}
			wg.Add(2)

			go func(wg *sync.WaitGroup) {
				p.watchEntityFiles()
				wg.Done()
			}(wg)

			go func(wg *sync.WaitGroup) {
				for i := range tc.create {
					f, err2 := os.Create(filepath.Join(dir, tc.create[i]))
					require.NoError(t, err2)

					_, err2 = f.Write([]byte(fmt.Sprintf("- type: \"file\"\n  id: \"%s\"\n", tc.create[i])))
					require.NoError(t, err2)

					f.Close()
				}

				time.Sleep(200 * time.Millisecond) // so the updates are processed, before we delete the same file :)

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

			assert.Nil(t, p.entityWatcher)

			p.mutex.RLock()
			assert.Empty(t, p.entityUpdates)
			assert.Empty(t, p.entityDeletes)
			p.mutex.RUnlock()

			var count int
			p.IterateEntities(func(models.Entity) {
				count++
			})
			assert.Equal(t, tc.want, count)
		})
	}
}
