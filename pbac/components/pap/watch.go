package pap

import (
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (p *pap) clearWatcher() {
	if p.watcher == nil {
		return
	}

	list := p.watcher.WatchList()
	for i := range list {
		_ = p.watcher.Remove(list[i])
	}
}

func (p *pap) watchFiles() {
	p.wTimer = time.NewTimer(watchTimerInterval)
	p.wTimer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			_ = p.watcher.Close()
			p.watcher = nil
			return

		case e := <-p.watcher.Events:
			// cache changes in a separate go-routine, so we keep the watcher-loop tight.
			go p.policyModified(e)

		case <-p.wTimer.C:
			// perform actual changes in separate go-routines, so we keep the watcher-loop tight.
			go p.processUpdates()
			go p.processDeletes()
		}
	}
}

func (p *pap) policyModified(e fsnotify.Event) {
	if p.wTimer != nil {
		p.wTimer.Reset(watchTimerInterval)
	}

	switch e.Op {
	case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
		p.mutex.Lock()
		p.updates = append(p.updates, e.Name)
		slices.Sort(p.updates)
		p.updates = slices.Compact(p.updates)
		p.mutex.Unlock()

	case fsnotify.Remove:
		p.mutex.Lock()
		p.deletes = append(p.deletes, e.Name)
		slices.Sort(p.deletes)
		p.deletes = slices.Compact(p.deletes)
		p.mutex.Unlock()

	default:
	}
}

func (p *pap) processUpdates() {
	for {
		var path string

		p.mutex.Lock()
		l := len(p.updates)
		if l > 0 {
			path = p.updates[0]
			p.updates = p.updates[1:]
		}
		p.mutex.Unlock()

		if l == 0 {
			return
		}

		if path != "" {
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				go p.processUpdate(path)
			}
		}
	}
}

func (p *pap) processUpdate(path string) {
	f, err2 := os.Open(path)
	if err2 != nil {
		return
	}
	defer f.Close()

	key := filepath.Base(path)

	p.mutex.RLock()
	_, ok := p.policies[key]
	p.mutex.RUnlock()

	if ok {
		_ = p.Replace(key, f)
	} else {
		_ = p.Add(key, f)
	}
}

func (p *pap) processDeletes() {
	for {
		var path string

		p.mutex.Lock()
		l := len(p.deletes)
		if l > 0 {
			path = p.deletes[0]
			p.deletes = p.deletes[1:]
		}
		p.mutex.Unlock()

		if l == 0 {
			return
		}

		if path != "" {
			_ = p.Remove(filepath.Base(path))
		}
	}
}

const watchTimerInterval = 100 * time.Millisecond
