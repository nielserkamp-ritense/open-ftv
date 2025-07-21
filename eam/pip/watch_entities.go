package pip

import (
	"os"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (p *PIP) clearEntityWatcher() {
	if p.entityWatcher == nil {
		return
	}

	list := p.entityWatcher.WatchList()
	for i := range list {
		_ = p.entityWatcher.Remove(list[i])
	}
}

func (p *PIP) watchEntityFiles() {
	p.entityTimer = time.NewTimer(watchTimerInterval)
	p.entityTimer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			_ = p.entityWatcher.Close()
			p.entityWatcher = nil
			return

		case e := <-p.entityWatcher.Events:
			// cache changes in a separate go-routine, so we keep the entityWatcher-loop tight.
			go p.entitiesModified(e)

		case <-p.entityTimer.C:
			// perform actual changes in separate go-routines, so we keep the entityWatcher-loop tight.
			go p.processEntityUpdates()
			go p.processEntityDeletes()
		}
	}
}

func (p *PIP) entitiesModified(e fsnotify.Event) {
	if p.entityTimer != nil {
		p.entityTimer.Reset(watchTimerInterval)
	}

	switch e.Op {
	case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
		p.mutex.Lock()
		p.entityUpdates = append(p.entityUpdates, e.Name)
		slices.Sort(p.entityUpdates)
		p.entityUpdates = slices.Compact(p.entityUpdates)
		p.mutex.Unlock()

	case fsnotify.Remove:
		p.mutex.Lock()
		p.entityDeletes = append(p.entityDeletes, e.Name)
		slices.Sort(p.entityDeletes)
		p.entityDeletes = slices.Compact(p.entityDeletes)
		p.mutex.Unlock()

	default:
	}
}

func (p *PIP) processEntityUpdates() {
	for {
		var path string

		p.mutex.Lock()
		l := len(p.entityUpdates)
		if l > 0 {
			path = p.entityUpdates[0]
			p.entityUpdates = p.entityUpdates[1:]
		}
		p.mutex.Unlock()

		if l == 0 {
			return
		}

		if path != "" {
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				// note that if one or more entities were removed from the file,
				// they will remain in memory. In most cases it will be safer to
				// just restart the service.
				go p.loadEntities(path)
			}
		}
	}
}

func (p *PIP) processEntityDeletes() {
	for {
		var path string

		p.mutex.Lock()
		l := len(p.entityDeletes)
		if l > 0 {
			path = p.entityDeletes[0]
			p.entityDeletes = p.entityDeletes[1:]
		}
		p.mutex.Unlock()

		if l == 0 {
			return
		}

		// TODO: implement
		// for now we just ignore deletes; it will be safer to restart the service,
		// as it some attributes from this file may have been replaced by another file.
		_ = path

	}
}
