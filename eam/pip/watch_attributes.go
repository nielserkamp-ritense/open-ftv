package pip

import (
	"os"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (p *pip) clearAttributeWatcher() {
	if p.attributeWatcher == nil {
		return
	}

	list := p.attributeWatcher.WatchList()
	for i := range list {
		_ = p.attributeWatcher.Remove(list[i])
	}
}

func (p *pip) watchAttributeFiles() {
	p.attributeTimer = time.NewTimer(watchTimerInterval)
	p.attributeTimer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			_ = p.attributeWatcher.Close()
			p.attributeWatcher = nil
			return

		case e := <-p.attributeWatcher.Events:
			// cache changes in a separate go-routine, so we keep the attributeWatcher-loop tight.
			go p.attributesModified(e)

		case <-p.attributeTimer.C:
			// perform actual changes in separate go-routines, so we keep the attributeWatcher-loop tight.
			go p.processAttributeUpdates()
			go p.processAttributeDeletes()
		}
	}
}

func (p *pip) attributesModified(e fsnotify.Event) {
	if p.attributeTimer != nil {
		p.attributeTimer.Reset(watchTimerInterval)
	}

	switch e.Op {
	case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
		p.mutex.Lock()
		p.attributeUpdates = append(p.attributeUpdates, e.Name)
		slices.Sort(p.attributeUpdates)
		p.attributeUpdates = slices.Compact(p.attributeUpdates)
		p.mutex.Unlock()

	case fsnotify.Remove:
		p.mutex.Lock()
		p.attributeDeletes = append(p.attributeDeletes, e.Name)
		slices.Sort(p.attributeDeletes)
		p.attributeDeletes = slices.Compact(p.attributeDeletes)
		p.mutex.Unlock()

	default:
	}
}

func (p *pip) processAttributeUpdates() {
	for {
		var path string

		p.mutex.Lock()
		l := len(p.attributeUpdates)
		if l > 0 {
			path = p.attributeUpdates[0]
			p.attributeUpdates = p.attributeUpdates[1:]
		}
		p.mutex.Unlock()

		if l == 0 {
			return
		}

		if path != "" {
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				// note that if one or more attributes were removed from the file,
				// they will remain in memory. In most cases it will be safer to
				// just restart the service.
				go p.loadAttributes(path)
			}
		}
	}
}

func (p *pip) processAttributeDeletes() {
	for {
		var path string

		p.mutex.Lock()
		l := len(p.attributeDeletes)
		if l > 0 {
			path = p.attributeDeletes[0]
			p.attributeDeletes = p.attributeDeletes[1:]
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

const watchTimerInterval = 100 * time.Millisecond
