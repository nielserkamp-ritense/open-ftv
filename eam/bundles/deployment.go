package bundles

import (
	"time"

	"github.com/goccy/go-json"

	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// NewDeployment instantiates a new bundle deployment.
func NewDeployment(version uint64, description string) *Deployment {
	now := time.Now().UTC()

	return &Deployment{
		version:     version,
		description: description,
		status:      Creating,
		created:     now,
		updated:     now,
	}
}

// Deployment represents the details of a bundle deployment.
type Deployment struct {
	version     uint64
	description string
	status      Status
	msg         string
	created     time.Time
	updated     time.Time
}

// Version returns the version number of the deployment.
func (d *Deployment) Version() uint64 {
	return d.version
}

// Status returns the status of the deployment.
func (d *Deployment) Status() Status {
	return d.status
}

// NextStatus updates the deployment status to the next level.
func (d *Deployment) NextStatus() Status {
	if d.status < Failed {
		d.status++
	}
	return d.status
}

// Failed marks the deployment as failed.
func (d *Deployment) Failed(msg string) Status {
	if d.status < Failed {
		d.status = Failed
		d.msg = msg
	}
	return d.status
}

// Completed marks the deployment as completed.
func (d *Deployment) Completed() Status {
	if d.status < Failed {
		d.status = Completed
	}
	return d.status
}

// MarshalJSON implements the json.Marshaler interface.
func (d *Deployment) MarshalJSON() ([]byte, error) {
	d2 := bundles2.Deployment{
		Version:     int(d.version),
		Description: d.description,
		Status:      int(d.status),
		Message:     d.msg,
		Created:     d.created.Format(time.RFC3339),
		Updated:     d.updated.Format(time.RFC3339),
	}
	return json.Marshal(d2)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Deployment) UnmarshalJSON(data []byte) (err error) {
	var d2 bundles2.Deployment
	if err = json.Unmarshal(data, &d2); err != nil {
		return
	}

	d.version = uint64(d2.Version)
	d.description = d2.Description
	d.status = Status(d2.Status)
	d.msg = d2.Message

	if d.created, err = time.Parse(time.RFC3339, d2.Created); err != nil {
		return
	}

	d.updated, err = time.Parse(time.RFC3339, d2.Updated)
	return
}
