package pap

import (
	"bytes"
	"os"
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
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	pol, err2 := NewPolicyFromStore(path, f)
	if err2 != nil {
		return
	}

	p.mutex.RLock()
	_, ok := p.policies[pol.ID()]
	p.mutex.RUnlock()

	if ok {
		_, _ = p.Replace(pol)
	} else {
		_, _ = p.Add(pol)
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

		pol, err2 := NewPolicyFromStore(path, bytes.NewReader([]byte{}))
		if err2 == nil {
			_, _ = p.Remove(pol.ID())
		}
	}
}

const watchTimerInterval = 100 * time.Millisecond
