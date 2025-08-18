package pap

import (
	"os"
	"time"

	"github.com/fsnotify/fsnotify"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (p *PAP) clearWatcher() {
	if p.watcher == nil {
		return
	}

	list := p.watcher.WatchList()
	for i := range list {
		_ = p.watcher.Remove(list[i])
	}
}

func (p *PAP) watchFiles() {
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

func (p *PAP) policyModified(e fsnotify.Event) {
	if p.wTimer != nil {
		p.wTimer.Reset(watchTimerInterval)
	}

	switch e.Op {
	case fsnotify.Remove:
		p.deployMutex.Lock()
		p.deletes[e.Name] = struct{}{}
		p.deployMutex.Unlock()

	default:
		p.deployMutex.Lock()
		p.updates[e.Name] = struct{}{}
		p.deployMutex.Unlock()
	}
}

func (p *PAP) processUpdates() {
	for {
		var path string

		p.deployMutex.Lock()
		for k := range p.updates {
			path = k
			delete(p.updates, k)
			break
		}
		p.deployMutex.Unlock()

		if path == "" {
			return
		}

		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			go p.processUpdate(path)
		}
	}
}

func (p *PAP) processUpdate(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	pol, err2 := models.NewPolicyFromStore(p.language, path, f)
	if err2 != nil {
		return
	}

	if prev, lastIndex, err3 := p.Read(pol.ID()); err3 == nil {
		_, _ = p.Update(prev, lastIndex, pol, storageUser)
	} else {
		_, _ = p.Create(pol, storageUser)
	}
}

func (p *PAP) processDeletes() {
	for {
		var path string

		p.deployMutex.Lock()
		for k := range p.deletes {
			path = k
			delete(p.deletes, k)
			break
		}
		p.deployMutex.Unlock()

		if path == "" {
			return
		}

		if prev, lastIndex, err := p.Read(path); err == nil {
			_, _ = p.Delete(prev, lastIndex, storageUser)
		}
	}
}

const (
	watchTimerInterval = 100 * time.Millisecond
	storageUser        = "*STORAGE*"
)
