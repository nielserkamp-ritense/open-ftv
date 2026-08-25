package management

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

var (
	// ErrNotFound reports that the addressed object does not exist. It is only returned to a
	// caller who was permitted the operation, so it never reveals an id to anyone else.
	ErrNotFound = errors.New("not found")
	// ErrExists reports that an object with the given id already exists.
	ErrExists = errors.New("already exists")
	// ErrInvalidTransition reports a status change the policy's life cycle does not allow.
	ErrInvalidTransition = errors.New("invalid status transition")
)

// PolicyService administers policies: every operation is one decision followed by one store call.
type PolicyService struct {
	logger  *slog.Logger
	store   *pap.PAP
	decider authorization.Decider
}

// PolicyDetails is a policy with the history the detail view shows next to it.
type PolicyDetails struct {
	Policy    *models.Policy
	AuditLog  []oas.AuditEntry
	UsageData []oas.UsageData
}

// NewPolicyService instantiates the policy service. A nil decider permits everything, for
// deployments and tests that run without authorization.
func NewPolicyService(logger *slog.Logger, store *pap.PAP, decider authorization.Decider) *PolicyService {
	return &PolicyService{logger: logger, store: store, decider: decider}
}

// List returns every policy the store holds.
func (s *PolicyService) List(ctx context.Context, caller *authorization.RequestPrincipal) ([]*models.Policy, error) {
	if err := s.decide(ctx, caller, models.ActionRead, policyCollection()); err != nil {
		return nil, err
	}

	return s.store.List("")
}

// Get returns one policy with its audit log and deployment usage.
func (s *PolicyService) Get(ctx context.Context, caller *authorization.RequestPrincipal, id string) (*PolicyDetails, error) {
	prev, _, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}

	if err := s.decide(ctx, caller, models.ActionRead, policyResource(id, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	out := &PolicyDetails{Policy: prev}

	if out.AuditLog, err = s.store.ReadAudit(id); err != nil {
		s.logger.Warn("failed to read policy audit", "id", id, "err", err)
	}

	if out.UsageData, err = s.store.ReadDeployments(id); err != nil {
		s.logger.Warn("failed to read policy deployments", "id", id, "err", err)
	}

	return out, nil
}

// Versions returns the version history of a policy; empty when the policy has none.
func (s *PolicyService) Versions(ctx context.Context, caller *authorization.RequestPrincipal, id string) (oas.PolicyVersions, error) {
	prev, _, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}

	if err := s.decide(ctx, caller, models.ActionRead, policyResource(id, prev)); err != nil {
		return nil, err
	}

	return s.store.ReadVersions(id)
}

// Version returns one version of a policy.
func (s *PolicyService) Version(ctx context.Context, caller *authorization.RequestPrincipal, id string, version int) (*oas.PolicyVersion, error) {
	prev, _, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}

	if err := s.decide(ctx, caller, models.ActionRead, policyResource(id, prev)); err != nil {
		return nil, err
	}

	out, err := s.store.ReadVersion(id, version)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, ErrNotFound
	}

	return out, nil
}

// Create adds a new policy. With forceUpsert an existing policy is updated instead, which is
// then decided as an update; without it an existing id is ErrExists.
func (s *PolicyService) Create(ctx context.Context, caller *authorization.RequestPrincipal, p *models.Policy, forceUpsert bool) (*models.Policy, error) {
	prev, lastIndex, err := s.store.Read(p.ID())
	if err != nil {
		return nil, err
	}

	switch {
	case prev == nil:
		if err := s.decide(ctx, caller, models.ActionCreate, policyResource(p.ID(), p)); err != nil {
			return nil, err
		}

		return s.store.Create(p, caller.Principal)

	case !forceUpsert:
		if err := s.decide(ctx, caller, models.ActionCreate, policyResource(p.ID(), prev)); err != nil {
			return nil, err
		}

		return nil, ErrExists

	default:
		if err := s.decide(ctx, caller, models.ActionUpdate, policyResource(p.ID(), prev)); err != nil {
			return nil, err
		}

		return s.store.Update(prev, lastIndex, p, caller.Principal)
	}
}

// Update replaces an existing policy. With forceUpsert a missing policy is created instead,
// which is then decided as a create; without it a missing id is ErrNotFound.
func (s *PolicyService) Update(ctx context.Context, caller *authorization.RequestPrincipal, p *models.Policy, forceUpsert bool) (*models.Policy, error) {
	prev, lastIndex, err := s.store.Read(p.ID())
	if err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		if err := s.decide(ctx, caller, models.ActionUpdate, policyResource(p.ID(), prev)); err != nil {
			return nil, err
		}

		return s.store.Update(prev, lastIndex, p, caller.Principal)

	case !forceUpsert:
		if err := s.decide(ctx, caller, models.ActionUpdate, policyResource(p.ID(), nil)); err != nil {
			return nil, err
		}

		return nil, ErrNotFound

	default:
		if err := s.decide(ctx, caller, models.ActionCreate, policyResource(p.ID(), p)); err != nil {
			return nil, err
		}

		return s.store.Create(p, caller.Principal)
	}
}

// SetStatus moves a policy to the given status: to accepted is the `accept` action, to
// deployed the `deploy` action, back to concept the `revert` action.
func (s *PolicyService) SetStatus(ctx context.Context, caller *authorization.RequestPrincipal, id string, status models.Status) (*models.Policy, error) {
	prev, lastIndex, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}

	if err := s.decide(ctx, caller, statusAction(status), policyResource(id, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	out, err := s.store.UpdateStatus(prev, lastIndex, status, caller.Principal)
	if errors.Is(err, pap.ErrInvalidStatusChange) {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTransition, err)
	}

	return out, err
}

// Restore makes an older version the current concept of the policy.
func (s *PolicyService) Restore(ctx context.Context, caller *authorization.RequestPrincipal, id string, version int) (*models.Policy, error) {
	prev, _, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}

	if err := s.decide(ctx, caller, models.ActionRestore, policyResource(id, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	return s.store.RestoreVersion(id, version, caller.Principal)
}

// Delete removes a policy. With ignoreMissing a missing id yields an empty policy carrying
// the id, so the call is idempotent; without it a missing id is ErrNotFound.
func (s *PolicyService) Delete(ctx context.Context, caller *authorization.RequestPrincipal, id string, ignoreMissing bool) (*models.Policy, error) {
	prev, lastIndex, err := s.store.Read(id)
	if err != nil {
		return nil, err
	}

	if err := s.decide(ctx, caller, models.ActionDelete, policyResource(id, prev)); err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		return s.store.Delete(prev, lastIndex, caller.Principal)
	case ignoreMissing:
		return models.NewPolicyFromData(id, "", "", "", &bytes.Buffer{})
	default:
		return nil, ErrNotFound
	}
}

// decide asks the PDP; without a decider every operation is permitted.
func (s *PolicyService) decide(ctx context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	if s.decider == nil {
		return nil
	}

	return s.decider.Decide(ctx, caller, action, resource)
}

// statusAction names the action a status change is decided as.
func statusAction(status models.Status) string {
	switch status {
	case models.StatusAccepted:
		return models.ActionAccept
	case models.StatusDeployed:
		return models.ActionDeploy
	default:
		return models.ActionRevert
	}
}

// policyCollection is the resource of a collection read.
func policyCollection() *models.Entity {
	return models.NewEntity(models.EntityTypePolicy, models.ResourceCollection, models.NewAttributeSet())
}

// policyResource is the resource of one policy, carrying the properties a policy may decide
// on. A nil object (not stored) contributes its id only.
func policyResource(id string, p *models.Policy) *models.Entity {
	attrs := models.NewAttributeSet()

	if p != nil {
		attrs.AddAttributeKV(models.AttrStatus, p.StatusName())
		attrs.AddAttributeKV(models.AttrTags, p.Tags())
		attrs.AddAttributeKV(models.AttrLanguage, p.Language())
	}

	return models.NewEntity(models.EntityTypePolicy, id, attrs)
}
