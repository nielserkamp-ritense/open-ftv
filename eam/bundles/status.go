package bundles

// Status represents the status of a bundle deployment.
type Status uint8

// List of deployment statuses.
const (
	Creating Status = iota + 1
	Gathering
	Merging
	Bundling
	Sending
	Failed
	Completed

	StatusMIN   = Creating
	StatusMAX   = Completed
	StatusCount = StatusMAX - StatusMIN + 1
)

// String implements the Stringer interface.
func (s Status) String() string {
	switch s {
	case Creating:
		return "Creating new deployment"
	case Gathering:
		return "Gathering policies & data"
	case Merging:
		return "Merging new deployment version in Git"
	case Bundling:
		return "Bundling policies & data"
	case Sending:
		return "Sending bundles"
	case Failed:
		return "Failed to complete successfully"
	case Completed:
		return "Finished successfully"
	default:
		return "???"
	}
}
