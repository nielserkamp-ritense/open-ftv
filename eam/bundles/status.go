package bundles

import "strings"

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

// Status returns a readable code for the status.
func (s Status) Status() string {
	switch s {
	case Creating:
		return "creating"
	case Gathering:
		return "gathering"
	case Merging:
		return "merging"
	case Bundling:
		return "bundling"
	case Sending:
		return "sending"
	case Failed:
		return "failed"
	case Completed:
		return "completed"
	default:
		return "???"
	}
}

// StatusFromString returns the status from the given code.
func StatusFromString(code string) Status {
	switch strings.ToLower(code) {
	case "creating":
		return Creating
	case "gathering":
		return Gathering
	case "merging":
		return Merging
	case "bundling":
		return Bundling
	case "sending":
		return Sending
	case "failed":
		return Failed
	case "completed":
		return Completed
	default:
		return 99
	}
}
