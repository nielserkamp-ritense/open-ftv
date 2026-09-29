package management

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AttributeStore is what the attribute service needs from the PIP: the attribute operations,
// without the rest. *pip.PIP satisfies it.
type AttributeStore interface {
	IterateAttributes(f models.AttributeIterator)
	GetAttribute(key string) (*models.Attribute, uint64, error)
	GetAttributeAudit(key string) ([]oas.AuditEntry, error)
	GetAttributeDeployments(key string) ([]oas.UsageData, error)
	GetAttributeVersions(key string) (oas.AttributeVersions, error)
	GetAttributeVersion(key string, version int) (*oas.AttributeVersion, error)
	AddAttributeFromOAS(in *oas.Attribute, user identity.Principal) (*models.Attribute, error)
	UpdateAttributeStatus(key string, status models.Status, user identity.Principal) (*models.Attribute, error)
	RestoreAttributeVersion(key string, version int, user identity.Principal) (*models.Attribute, error)
	RemoveAttribute(key string, user identity.Principal) (*models.Attribute, error)
}

// AttributeService administers attributes, the PIP's key/value data: every operation is one
// decision followed by one store call.
type AttributeService struct {
	logger     *slog.Logger
	store      AttributeStore
	authorizer authorization.Authorizer
}

// AttributeDetails is an attribute with the history the detail view shows next to it.
type AttributeDetails struct {
	Attribute *models.Attribute
	AuditLog  []oas.AuditEntry
	UsageData []oas.UsageData
}

// NewAttributeService instantiates the attribute service. A nil authorizer permits everything,
// for deployments and tests that run without authorization.
func NewAttributeService(logger *slog.Logger, store AttributeStore, authorizer authorization.Authorizer) *AttributeService {
	return &AttributeService{logger: logger, store: store, authorizer: authorizer}
}

// List returns every attribute the store holds.
func (s *AttributeService) List(ctx context.Context, caller *authorization.RequestPrincipal) ([]*models.Attribute, error) {
	if err := s.authorize(ctx, caller, models.ActionRead, attributeCollection()); err != nil {
		return nil, err
	}

	var out []*models.Attribute

	s.store.IterateAttributes(func(a *models.Attribute) { out = append(out, a) })

	return out, nil
}

// Get returns one attribute with its audit log and deployment usage.
func (s *AttributeService) Get(ctx context.Context, caller *authorization.RequestPrincipal, key string) (*AttributeDetails, error) { //nolint:dupl // the attribute and entity services are the same shape over different types
	prev, err := s.load(ctx, caller, models.ActionRead, key)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, attributeResource(key, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	out := &AttributeDetails{Attribute: prev}

	if out.AuditLog, err = s.store.GetAttributeAudit(key); err != nil {
		s.logger.Warn("failed to read attribute audit", "key", key, "err", err)
	}

	if out.UsageData, err = s.store.GetAttributeDeployments(key); err != nil {
		s.logger.Warn("failed to read attribute deployments", "key", key, "err", err)
	}

	return out, nil
}

// Versions returns the version history of an attribute; empty when it has none.
func (s *AttributeService) Versions(ctx context.Context, caller *authorization.RequestPrincipal, key string) (oas.AttributeVersions, error) {
	prev, err := s.load(ctx, caller, models.ActionRead, key)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, attributeResource(key, prev)); err != nil {
		return nil, err
	}

	return s.store.GetAttributeVersions(key)
}

// Version returns one version of an attribute.
func (s *AttributeService) Version(ctx context.Context, caller *authorization.RequestPrincipal, key string, version int) (*oas.AttributeVersion, error) {
	prev, err := s.load(ctx, caller, models.ActionRead, key)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, attributeResource(key, prev)); err != nil {
		return nil, err
	}

	out, err := s.store.GetAttributeVersion(key, version)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, ErrNotFound
	}

	return out, nil
}

// Create adds a new attribute. With forceUpsert an existing attribute is updated instead, which
// is then decided as an update; without it an existing key is ErrExists. The decision on a
// create carries the submitted properties, the decision on an update the stored ones.
func (s *AttributeService) Create(ctx context.Context, caller *authorization.RequestPrincipal, in *oas.Attribute, forceUpsert bool) (*models.Attribute, error) {
	prev, err := s.load(ctx, caller, models.ActionCreate, in.Key)
	if err != nil {
		return nil, err
	}

	switch {
	case prev == nil:
		if err := s.authorize(ctx, caller, models.ActionCreate, submittedAttribute(in)); err != nil {
			return nil, err
		}

		return s.store.AddAttributeFromOAS(in, caller.Attribution())

	case !forceUpsert:
		if err := s.authorize(ctx, caller, models.ActionCreate, attributeResource(in.Key, prev)); err != nil {
			return nil, err
		}

		return nil, ErrExists

	default:
		if err := s.authorize(ctx, caller, models.ActionUpdate, attributeResource(in.Key, prev)); err != nil {
			return nil, err
		}

		return s.store.AddAttributeFromOAS(in, caller.Attribution())
	}
}

// Update replaces an existing attribute. With forceUpsert a missing attribute is created instead,
// which is then decided as a create; without it a missing key is ErrNotFound.
func (s *AttributeService) Update(ctx context.Context, caller *authorization.RequestPrincipal, in *oas.Attribute, forceUpsert bool) (*models.Attribute, error) {
	prev, err := s.load(ctx, caller, models.ActionUpdate, in.Key)
	if err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		if err := s.authorize(ctx, caller, models.ActionUpdate, attributeResource(in.Key, prev)); err != nil {
			return nil, err
		}

		return s.store.AddAttributeFromOAS(in, caller.Attribution())

	case !forceUpsert:
		if err := s.authorize(ctx, caller, models.ActionUpdate, attributeResource(in.Key, nil)); err != nil {
			return nil, err
		}

		return nil, ErrNotFound

	default:
		if err := s.authorize(ctx, caller, models.ActionCreate, submittedAttribute(in)); err != nil {
			return nil, err
		}

		return s.store.AddAttributeFromOAS(in, caller.Attribution())
	}
}

// SetStatus moves an attribute to the given status: to accepted is the `accept` action, back
// to concept the `revert` action. The PIP allows no other transition, so any other status is
// refused as input before anything is decided.
func (s *AttributeService) SetStatus(ctx context.Context, caller *authorization.RequestPrincipal, key string, status models.Status) (*models.Attribute, error) {
	if err := pipStatus(status); err != nil {
		return nil, err
	}

	prev, err := s.load(ctx, caller, statusAction(status), key)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, statusAction(status), attributeResource(key, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	out, err := s.store.UpdateAttributeStatus(key, status, caller.Attribution())
	if errors.Is(err, pip.ErrInvalidStatusChange) {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTransition, err)
	}

	return out, err
}

// Restore makes an older version the current concept of the attribute.
func (s *AttributeService) Restore(ctx context.Context, caller *authorization.RequestPrincipal, key string, version int) (*models.Attribute, error) {
	prev, err := s.load(ctx, caller, models.ActionRestore, key)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRestore, attributeResource(key, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	// The version is looked up before the restore: the store reports a missing one as nil,
	// and would otherwise restore from nothing.
	old, err := s.store.GetAttributeVersion(key, version)
	if err != nil {
		return nil, err
	}

	if old == nil {
		return nil, ErrNotFound
	}

	return s.store.RestoreAttributeVersion(key, version, caller.Attribution())
}

// Delete removes an attribute. With ignoreMissing a missing key yields an empty attribute
// carrying the key, so the call is idempotent; without it a missing key is ErrNotFound.
func (s *AttributeService) Delete(ctx context.Context, caller *authorization.RequestPrincipal, key string, ignoreMissing bool) (*models.Attribute, error) {
	prev, err := s.load(ctx, caller, models.ActionDelete, key)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionDelete, attributeResource(key, prev)); err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		return s.store.RemoveAttribute(key, caller.Attribution())
	case ignoreMissing:
		return models.NewAttribute(key, nil), nil
	default:
		return nil, ErrNotFound
	}
}

// load reads the attribute an operation acts on. A read that fails is still decided, on the
// key-only resource, and the decision is reported first, so the store's state and its error
// text reach permitted callers only.
func (s *AttributeService) load(ctx context.Context, caller *authorization.RequestPrincipal, action, key string) (*models.Attribute, error) {
	prev, _, err := s.store.GetAttribute(key)
	if err == nil {
		return prev, nil
	}

	if derr := s.authorize(ctx, caller, action, attributeResource(key, nil)); derr != nil {
		return nil, derr
	}

	return nil, err
}

// authorize asks the PDP; without an authorizer every operation is permitted.
func (s *AttributeService) authorize(ctx context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	return decide(ctx, s.authorizer, caller, action, resource)
}

// attributeCollection is the resource of a collection read.
func attributeCollection() *models.Entity {
	return models.NewEntity(models.EntityTypeAttribute, models.ResourceCollection, models.NewAttributeSet())
}

// attributeResource is the resource of one stored attribute, carrying the properties a policy
// may decide on. A nil object (not stored) contributes its key only.
func attributeResource(key string, a *models.Attribute) *models.Entity {
	if a == nil {
		return pipResource(models.EntityTypeAttribute, key, "", nil)
	}

	return pipResource(models.EntityTypeAttribute, key, a.StatusName(), a.Tags())
}

// submittedAttribute is the resource of an attribute that is not stored yet, carrying the
// submitted properties. The PIP stores a new attribute as concept whatever the body says.
func submittedAttribute(in *oas.Attribute) *models.Entity {
	return pipResource(models.EntityTypeAttribute, in.Key, models.StatusConcept.String(), in.Metadata.Tags)
}

// pipStatus refuses a status PIP data cannot take: it is concept or accepted, never deployed
// through a status change (a deployment ships it).
func pipStatus(status models.Status) error {
	if status != models.StatusConcept && status != models.StatusAccepted {
		return fmt.Errorf("%w: PIP data is %s or %s, not %s", ErrInvalidTransition, models.StatusConcept.String(), models.StatusAccepted.String(), status.String())
	}

	return nil
}

// pipResource builds the resource of a PIP object: status and tags when known, nothing more
// until a policy needs it (ADR 0006).
func pipResource(tp, id, status string, tags []string) *models.Entity {
	attrs := models.NewAttributeSet()

	if status != "" {
		attrs.AddAttributeKV(models.AttrStatus, status)
		attrs.AddAttributeKV(models.AttrTags, tags)
	}

	return models.NewEntity(tp, id, attrs)
}
