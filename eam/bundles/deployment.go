package bundles

import (
	"time"

	"github.com/goccy/go-json"

	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// NewDeployment instantiates a new bundle deployment.
func NewDeployment(version uint64, title, description, user string) *Deployment {
	now := time.Now().UTC()

	return &Deployment{
		version:     version,
		title:       title,
		description: description,
		status:      Creating,
		created:     now,
		createdBy:   user,
		updated:     now,
		updatedBy:   user,
	}
}

// Deployment represents the details of a bundle deployment.
type Deployment struct {
	version     uint64
	title       string
	description string
	status      Status
	msg         string
	created     time.Time
	createdBy   string
	updated     time.Time
	updatedBy   string
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
		d.updated = time.Now().UTC()
		d.updatedBy = "*SYSTEM*"
	}
	return d.status
}

// Completed marks the deployment as completed.
func (d *Deployment) Completed() Status {
	if d.status < Failed {
		d.status = Completed
		d.updated = time.Now().UTC()
		d.updatedBy = "*SYSTEM*"
	}
	return d.status
}

// ToOAS returns the deployment as an OAS model.
func (d *Deployment) ToOAS() *oas.Deployment {
	return &oas.Deployment{
		Version:     int(d.version),
		Title:       d.title,
		Description: d.description,
		Status:      int(d.status),
		Message:     d.msg,
		Audit: oas.ObjectAudit{
			Created:   d.created.Format(time.RFC3339),
			CreatedBy: oas.Principal{Id: d.createdBy},
			Updated:   d.updated.Format(time.RFC3339),
			UpdatedBy: optionalPrincipal(d.updatedBy),
		},
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (d *Deployment) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.ToOAS())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Deployment) UnmarshalJSON(data []byte) (err error) {
	var d2 oas.Deployment
	if err = json.Unmarshal(data, &d2); err != nil {
		return
	}

	d.version = uint64(d2.Version)
	d.title = d2.Title
	d.description = d2.Description
	d.status = Status(d2.Status)
	d.msg = d2.Message
	d.createdBy = d2.Audit.CreatedBy.Id
	d.updatedBy = principalID(d2.Audit.UpdatedBy)

	if d.created, err = time.Parse(time.RFC3339, d2.Audit.Created); err != nil {
		return
	}

	d.updated, err = time.Parse(time.RFC3339, d2.Audit.Updated)
	return
}

// optionalPrincipal returns nil for an absent attribution, so the field is omitted from the
// response rather than serialized as a principal identifying nobody.
func optionalPrincipal(id string) *oas.Principal {
	if id == "" {
		return nil
	}

	return &oas.Principal{Id: id}
}

// principalID reads the id from an optional attribution, which is absent when nothing set it.
func principalID(p *oas.Principal) string {
	if p == nil {
		return ""
	}

	return p.Id
}
