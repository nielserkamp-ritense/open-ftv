package bundles

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// StatusHandler represents the interface for managing the last deployment status.
type StatusHandler interface {
	Advance() (*Deployment, error)
	Fail(msg string) (*Deployment, error)
	CreateBundleAudit(ctx context.Context, version uint64, cfg *Config, bundle *Bundle) error
}

// Run executes the required steps for a new deployment.
//
// The various stages of a run are:
// - *Creating*
// - *Gathering*
// - *Merging*
// - *Bundling*
// - *Sending*
//
// When an unrecoverable error is encountered, the status of the deployment will be set to *Failed*.
// On successful completion of the run, the status will be set to *Completed*.
//
// Each stage is designed to be restartable.
// After a broken run (e.g., due to the service being restarted), the service must call this function during initialization.
// Run will determine the status of the given deployment, and automatically continue where it was interrupted.
// If the status is *Failed* or *Completed*, Run will do nothing.
func (m *Manager) Run(d *Deployment, handler StatusHandler) any {
	// The context is canceled when the deployment status reaches *Failed* or *Completed*.
	ctx, cancel := context.WithCancel(m.ctx)

	r := &runner{
		ctx:           ctx,
		cancel:        cancel,
		m:             m,
		handler:       handler,
		d:             d,
		logger:        m.logger,
		bundleTimeout: m.bundleTimeout,
		client:        m.client,
	}

	go r.run() // asynchronous !
	return r
}

type runner struct {
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *slog.Logger
	m             *Manager
	d             *Deployment
	bundleTimeout time.Duration
	client        *http.Client
	handler       StatusHandler
	bundleCount   uint64
	targetCount   uint64
	bundledCount  uint64
	sendCount     uint64
	gitHash       string
	bundles       map[string]*Bundle
	targets       map[string]*sendJob
	policies      []*models.Policy
	attributes    []*models.Attribute
	entities      []*models.Entity
	relations     []*models.Relation
}

// This function should be called in a separate go-routine.
//
// It is designed to be called at the start of a deployment run and after each status update.
func (r *runner) run() {
	if r.ctx.Err() != nil {
		return
	}

	switch r.d.status {
	case Creating:
		r.creating()
	case Gathering:
		r.gathering()
	case Merging:
		r.merging()
	case Bundling:
		r.bundling()
	case Sending:
		r.sending()
	case Failed, Completed:
		r.cancel()
	default:
		r.logger.Error("bundle-runner: invalid deployment status", "status", r.d.status.String())
	}
}

func (r *runner) debug(msg string, params ...any) {
	m, p := r.prepareLog(msg, params)
	r.logger.Debug(m, p...)
}

func (r *runner) info(msg string, params ...any) {
	m, p := r.prepareLog(msg, params)
	r.logger.Info(m, p...)
}

func (r *runner) warn(msg string, params ...any) {
	m, p := r.prepareLog(msg, params)
	r.logger.Warn(m, p...)
}

func (r *runner) error(msg string, params ...any) {
	m, p := r.prepareLog(msg, params)
	r.logger.Error(m, p...)
}

func (r *runner) prepareLog(msg string, params []any) (string, []any) {
	return fmt.Sprintf("bundle-runner: %s", msg),
		append([]any{"version", r.d.version, "title", r.d.title}, params...)
}

// The advance() function advances the status to the next stage.
//
// If an error is encountered when updating the status, it will try to set the status to *Failed*.
// Any error encountered during this process, will force the function to return false.
//
// If the status is updated successfully, the function will start the next stage (if necessary) and return true.
func (r *runner) advance(stage string) bool {
	d, err := r.handler.Advance()
	if err != nil {
		r.error(fmt.Sprintf("%s stage failed to advance status", stage), "error", err)

		if d, err = r.handler.Fail(err.Error()); err != nil {
			r.error(fmt.Sprintf("%s stage unable to set fail status", stage), "error", err)
		} else {
			r.d = d
		}

		r.cancel()
		return false
	}

	r.d = d
	r.debug(fmt.Sprintf("%s stage completed successfully", stage))

	if r.d.status != Completed {
		if r.m.stageDelay > 0 {
			// force delay as configured.
			time.Sleep(r.m.stageDelay)
		}

		go r.run() // run the next stage.
	}

	return true
}
