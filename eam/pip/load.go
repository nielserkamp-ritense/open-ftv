package pip

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func (p *PIP) loadFromStore() {
	if p.attrStore != "" {
		p.clearAttributeWatcher()
		p.iterateFolders(p.attrStore, p.recurse, p.attributeWatcher, p.loadAttributes)
	}
	if p.entityStore != "" {
		p.clearEntityWatcher()
		p.iterateFolders(p.entityStore, p.recurse, p.entityWatcher, p.loadEntities)
	}
}

func (p *PIP) iterateFolders(path string, recurse bool, watcher *fsnotify.Watcher, f func(path string)) {
	err := filepath.WalkDir(path, func(path2 string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if recurse || path2 == path {
				if watcher != nil {
					_ = watcher.Add(path)
				}
				return nil
			}
			return filepath.SkipDir
		}

		if base := filepath.Base(path2); !strings.HasPrefix(base, ".") {
			f(path2)
		}
		return nil
	})

	if err != nil {
		p.logger.Error("pip: error iterating folders", "path", path, "err", err)
	}
}
