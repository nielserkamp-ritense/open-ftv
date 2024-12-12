package pip

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/fds/ledenlijst"
)

// MaturityLevel represents the maturity level of an FDS member organization.
type MaturityLevel int8

// FDS represents the interface to retrieve attributes from a specific "Federatief Data Stelsel".
type FDS interface {
	IsMember(oin string) bool               // indicates if an organization (by OIN) is a member of this FDS.
	MaturityLevel(oin string) MaturityLevel // returns the maturity level of a member organization of this FDS, or -1 for non-members.
}

// IsMember implements the FDS interface.
func (f *fds) IsMember(oin string) bool {
	f.mutex.RLock()
	_, ok := f.leden[oin]
	f.mutex.RUnlock()
	return ok
}

// MaturityLevel implements the FDS interface.
func (f *fds) MaturityLevel(oin string) MaturityLevel {
	f.mutex.RLock()
	level, ok := f.leden[oin]
	f.mutex.RUnlock()

	if ok {
		return level
	}
	return -1
}

func newFDS(ctx context.Context, logger *slog.Logger, u *url.URL, ttl time.Duration) FDS {
	if ctx == nil {
		ctx = context.Background()
	}

	f := &fds{
		ctx:    ctx,
		logger: logger,
		url:    u,
		ttl:    ttl,
		leden:  make(map[string]MaturityLevel),
		mutex:  sync.RWMutex{},
	}

	go f.refresh() // initial refresh.
	go f.run()     // background refresher (based on TTL).

	return f
}

func (f *fds) refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := f.url.String()
	req, err := http.NewRequestWithContext(ctx, fiber.MethodGet, u, nil)
	if err != nil {
		f.logger.Error("failed to create FDS ledenlijst request", "url", u, "error", err)
		return
	}

	resp, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		f.logger.Error("failed to retrieve FDS ledenlijst", "url", u, "error", err2)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		f.logger.Error("invalid return code FDS ledenlijst", "url", u, "code", resp.StatusCode, "status", resp.Status)
		return
	}

	list := make([]ledenlijst.Organization, 0, 32)
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		f.logger.Error("invalid response from FDS ledenlijst", "url", u, "error", err)
		return
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	clear(f.leden)

	for i := range list {
		org := list[i]
		if member, ok := org.Attributes["isMember"].(bool); ok && member {
			var maturity int
			if n, ok2 := org.Attributes["maturity"].(float64); ok2 {
				maturity = int(n)
			}
			f.leden[org.Oin] = MaturityLevel(maturity)
		}
	}

	// go f.logger.Info("FDS ledenlijst updated", "url", u, "count", len(list))
}

func (f *fds) run() {
	t := time.NewTicker(f.ttl)

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-t.C:
			go f.refresh()
		}
	}
}

type fds struct {
	ctx    context.Context
	logger *slog.Logger
	url    *url.URL
	ttl    time.Duration
	leden  map[string]MaturityLevel
	mutex  sync.RWMutex
}
