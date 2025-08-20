package pap

import (
	"os"
	"time"

	"github.com/fsnotify/fsnotify"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (p *PAP) clearWatcher() {
	if p.policyWatcher == nil {
		return
	}

	list := p.policyWatcher.WatchList()
	for i := range list {
		_ = p.policyWatcher.Remove(list[i])
	}
}

func (p *PAP) watchFiles() {
	p.policyTimer = time.NewTimer(watchTimerInterval)
	p.policyTimer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			_ = p.policyWatcher.Close()
			p.policyWatcher = nil
			return

		case e := <-p.policyWatcher.Events:
			// cache changes in a separate go-routine, so we keep the watcher-loop tight.
			go p.policyModified(e)

		case <-p.policyTimer.C:
			// perform actual changes in separate go-routines, so we keep the watcher-loop tight.
			go p.processUpdates()
			go p.processDeletes()
		}
	}
}

func (p *PAP) policyModified(e fsnotify.Event) {
	if p.policyTimer != nil {
		p.policyTimer.Reset(watchTimerInterval)
	}

	switch e.Op {
	case fsnotify.Remove:
		p.deployMutex.Lock()
		p.policyDeletes[e.Name] = struct{}{}
		p.deployMutex.Unlock()

	default:
		p.deployMutex.Lock()
		p.policyUpdates[e.Name] = struct{}{}
		p.deployMutex.Unlock()
	}
}

func (p *PAP) processUpdates() {
	for {
		var path string

		p.deployMutex.Lock()
		for k := range p.policyUpdates {
			path = k
			delete(p.policyUpdates, k)
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

	if prev, lastIndex, err3 := p.Read(pol.ID()); err3 == nil && prev != nil {
		_, _ = p.Update(prev, lastIndex, pol, storageUser)
	} else {
		_, _ = p.Create(pol, storageUser)
	}
}

func (p *PAP) processDeletes() {
	for {
		var path string

		p.deployMutex.Lock()
		for k := range p.policyDeletes {
			path = k
			delete(p.policyDeletes, k)
			break
		}
		p.deployMutex.Unlock()

		if path == "" {
			return
		}

		if prev, lastIndex, err := p.Read(path); err == nil && prev != nil {
			_, _ = p.Delete(prev, lastIndex, storageUser)
		}
	}
}

const (
	watchTimerInterval = 100 * time.Millisecond
	storageUser        = "*STORAGE*"
)
