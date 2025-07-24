package bundles

// StatusHandler represents the interface for managing the last deployment status.
type StatusHandler interface {
	Advance() (*Deployment, error)
	Fail(msg string) (*Deployment, error)
}

// Run executes a new deployment from start to finish.
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
// After a broken run (due to the service being restarted) the service must call this function.
// Run will determine the status of the given deployment, and automatically continue where it was interrupted,
// as long as the status is not *Failed* or *Completed*.
func (m *Manager) Run(d *Deployment, handler StatusHandler) {
	r := &runner{m: m, handler: handler, d: d}
	r.run()
}

type runner struct {
	m       *Manager
	d       *Deployment
	handler StatusHandler
}

func (r *runner) run() {

}
