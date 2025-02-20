package pap

import (
	"bytes"
	"os"
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
	case fsnotify.Remove:
		p.mutex.Lock()
		p.deletes[e.Name] = struct{}{}
		p.mutex.Unlock()

	default:
		p.mutex.Lock()
		p.updates[e.Name] = struct{}{}
		p.mutex.Unlock()
	}
}

func (p *pap) processUpdates() {
	for {
		var path string

		p.mutex.Lock()
		for k := range p.updates {
			path = k
			delete(p.updates, k)
			break
		}
		p.mutex.Unlock()

		if path == "" {
			return
		}

		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			go p.processUpdate(path)
		}
	}
}

func (p *pap) processUpdate(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	pol, err2 := NewPolicyFromStore(p.language, path, f)
	if err2 != nil {
		return
	}

	if prev, err3 := p.Read(pol.Language(), pol.ID()); err3 == nil {
		_, _ = p.Update(prev, pol)
	} else {
		_, _ = p.Create(pol)
	}
}

func (p *pap) processDeletes() {
	for {
		var path string

		p.mutex.Lock()
		for k := range p.deletes {
			path = k
			delete(p.deletes, k)
			break
		}
		p.mutex.Unlock()

		if path == "" {
			return
		}

		if pol, err2 := NewPolicyFromStore(p.language, path, bytes.NewReader([]byte{})); err2 == nil {
			_, _ = p.Delete(pol)
		}
	}
}

const watchTimerInterval = 100 * time.Millisecond
