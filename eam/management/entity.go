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

// EntityStore is what the entity service needs from the PIP: the entity operations, without
// the rest. *pip.PIP satisfies it. Entities are addressed by their UID, `type::id`.
type EntityStore interface {
	IterateEntities(f models.EntityIterator)
	GetEntity(uid string) (*models.Entity, uint64, error)
	GetEntityAudit(uid string) ([]oas.AuditEntry, error)
	GetEntityDeployments(uid string) ([]oas.UsageData, error)
	GetEntityVersions(uid string) (oas.EntityVersions, error)
	GetEntityVersion(uid string, version int) (*oas.EntityVersion, error)
	AddEntityFromOAS(in *oas.Entity, user identity.Principal) (*models.Entity, error)
	UpdateEntityStatus(uid string, status models.Status, user identity.Principal) (*models.Entity, error)
	RestoreEntityVersion(uid string, version int, user identity.Principal) (*models.Entity, error)
	RemoveEntity(uid string, user identity.Principal) (*models.Entity, error)
}

// EntityService administers entities, the PIP's typed objects: every operation is one decision
// followed by one store call. The management-plane resource of an entity is `entity::<type::id>`.
type EntityService struct {
	logger     *slog.Logger
	store      EntityStore
	authorizer authorization.Authorizer
}

// EntityDetails is an entity with the history the detail view shows next to it.
type EntityDetails struct {
	Entity    *models.Entity
	AuditLog  []oas.AuditEntry
	UsageData []oas.UsageData
}

// NewEntityService instantiates the entity service. A nil authorizer permits everything, for
// deployments and tests that run without authorization.
func NewEntityService(logger *slog.Logger, store EntityStore, authorizer authorization.Authorizer) *EntityService {
	return &EntityService{logger: logger, store: store, authorizer: authorizer}
}

// List returns every entity the store holds.
func (s *EntityService) List(ctx context.Context, caller *authorization.RequestPrincipal) ([]*models.Entity, error) {
	if err := s.authorize(ctx, caller, models.ActionRead, entityCollection()); err != nil {
		return nil, err
	}

	var out []*models.Entity

	s.store.IterateEntities(func(e *models.Entity) { out = append(out, e) })

	return out, nil
}

// Get returns one entity with its audit log and deployment usage.
func (s *EntityService) Get(ctx context.Context, caller *authorization.RequestPrincipal, uid string) (*EntityDetails, error) { //nolint:dupl // the attribute and entity services are the same shape over different types
	prev, err := s.load(ctx, caller, models.ActionRead, uid)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, entityResource(uid, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	out := &EntityDetails{Entity: prev}

	if out.AuditLog, err = s.store.GetEntityAudit(uid); err != nil {
		s.logger.Warn("failed to read entity audit", "uid", uid, "err", err)
	}

	if out.UsageData, err = s.store.GetEntityDeployments(uid); err != nil {
		s.logger.Warn("failed to read entity deployments", "uid", uid, "err", err)
	}

	return out, nil
}

// Versions returns the version history of an entity; empty when it has none.
func (s *EntityService) Versions(ctx context.Context, caller *authorization.RequestPrincipal, uid string) (oas.EntityVersions, error) {
	prev, err := s.load(ctx, caller, models.ActionRead, uid)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, entityResource(uid, prev)); err != nil {
		return nil, err
	}

	return s.store.GetEntityVersions(uid)
}

// Version returns one version of an entity.
func (s *EntityService) Version(ctx context.Context, caller *authorization.RequestPrincipal, uid string, version int) (*oas.EntityVersion, error) {
	prev, err := s.load(ctx, caller, models.ActionRead, uid)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, entityResource(uid, prev)); err != nil {
		return nil, err
	}

	out, err := s.store.GetEntityVersion(uid, version)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, ErrNotFound
	}

	return out, nil
}

// Create adds a new entity. With forceUpsert an existing entity is updated instead, which is
// then decided as an update; without it an existing uid is ErrExists. The decision on a create
// carries the submitted properties, the decision on an update the stored ones. The status in
// the body is ignored: a new entity is a concept and an update keeps the stored status, so a
// status changes through SetStatus only (review, SCRUM-145).
func (s *EntityService) Create(ctx context.Context, caller *authorization.RequestPrincipal, in *oas.Entity, forceUpsert bool) (*models.Entity, error) {
	uid := models.EntityUID(in.Type, in.Id)

	prev, err := s.load(ctx, caller, models.ActionCreate, uid)
	if err != nil {
		return nil, err
	}

	switch {
	case prev == nil:
		if err := s.authorize(ctx, caller, models.ActionCreate, submittedEntity(in)); err != nil {
			return nil, err
		}

		return s.store.AddEntityFromOAS(asConcept(in), caller.Attribution())

	case !forceUpsert:
		if err := s.authorize(ctx, caller, models.ActionCreate, entityResource(uid, prev)); err != nil {
			return nil, err
		}

		return nil, ErrExists

	default:
		if err := s.authorize(ctx, caller, models.ActionUpdate, entityResource(uid, prev)); err != nil {
			return nil, err
		}

		return s.store.AddEntityFromOAS(withStatus(in, prev), caller.Attribution())
	}
}

// Update replaces an existing entity. With forceUpsert a missing entity is created instead,
// which is then decided as a create; without it a missing uid is ErrNotFound.
func (s *EntityService) Update(ctx context.Context, caller *authorization.RequestPrincipal, in *oas.Entity, forceUpsert bool) (*models.Entity, error) {
	uid := models.EntityUID(in.Type, in.Id)

	prev, err := s.load(ctx, caller, models.ActionUpdate, uid)
	if err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		if err := s.authorize(ctx, caller, models.ActionUpdate, entityResource(uid, prev)); err != nil {
			return nil, err
		}

		return s.store.AddEntityFromOAS(withStatus(in, prev), caller.Attribution())

	case !forceUpsert:
		if err := s.authorize(ctx, caller, models.ActionUpdate, entityResource(uid, nil)); err != nil {
			return nil, err
		}

		return nil, ErrNotFound

	default:
		if err := s.authorize(ctx, caller, models.ActionCreate, submittedEntity(in)); err != nil {
			return nil, err
		}

		return s.store.AddEntityFromOAS(asConcept(in), caller.Attribution())
	}
}

// SetStatus moves an entity to the given status: to accepted is the `accept` action, back to
// concept the `revert` action. The PIP allows no other transition, so any other status is
// refused as input before anything is decided.
func (s *EntityService) SetStatus(ctx context.Context, caller *authorization.RequestPrincipal, uid string, status models.Status) (*models.Entity, error) {
	if err := pipStatus(status); err != nil {
		return nil, err
	}

	prev, err := s.load(ctx, caller, statusAction(status), uid)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, statusAction(status), entityResource(uid, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	out, err := s.store.UpdateEntityStatus(uid, status, caller.Attribution())
	if errors.Is(err, pip.ErrInvalidStatusChange) {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTransition, err)
	}

	return out, err
}

// Restore makes an older version the current concept of the entity.
func (s *EntityService) Restore(ctx context.Context, caller *authorization.RequestPrincipal, uid string, version int) (*models.Entity, error) {
	prev, err := s.load(ctx, caller, models.ActionRestore, uid)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRestore, entityResource(uid, prev)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	// The version is looked up before the restore: the store reports a missing one as nil,
	// and would otherwise restore from nothing.
	old, err := s.store.GetEntityVersion(uid, version)
	if err != nil {
		return nil, err
	}

	if old == nil {
		return nil, ErrNotFound
	}

	return s.store.RestoreEntityVersion(uid, version, caller.Attribution())
}

// Delete removes an entity. With ignoreMissing a missing uid yields an empty entity carrying
// the type and id, so the call is idempotent; without it a missing uid is ErrNotFound.
func (s *EntityService) Delete(ctx context.Context, caller *authorization.RequestPrincipal, uid string, ignoreMissing bool) (*models.Entity, error) {
	prev, err := s.load(ctx, caller, models.ActionDelete, uid)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionDelete, entityResource(uid, prev)); err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		return s.store.RemoveEntity(uid, caller.Attribution())
	case ignoreMissing:
		ns, id := models.SplitEntityUID(uid)
		return models.NewEntity(ns, id, models.NewAttributeSet()), nil
	default:
		return nil, ErrNotFound
	}
}

// load reads the entity an operation acts on. A read that fails is still decided, on the
// uid-only resource, and the decision is reported first, so the store's state and its error
// text reach permitted callers only.
func (s *EntityService) load(ctx context.Context, caller *authorization.RequestPrincipal, action, uid string) (*models.Entity, error) {
	prev, _, err := s.store.GetEntity(uid)
	if err == nil {
		return prev, nil
	}

	if derr := s.authorize(ctx, caller, action, entityResource(uid, nil)); derr != nil {
		return nil, derr
	}

	return nil, err
}

// authorize asks the PDP; without an authorizer every operation is permitted.
func (s *EntityService) authorize(ctx context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	return decide(ctx, s.authorizer, caller, action, resource)
}

// entityCollection is the resource of a collection read.
func entityCollection() *models.Entity {
	return models.NewEntity(models.EntityTypeEntity, models.ResourceCollection, models.NewAttributeSet())
}

// entityResource is the resource of one stored entity, carrying the properties a policy may
// decide on. A nil object (not stored) contributes its uid only.
func entityResource(uid string, e *models.Entity) *models.Entity {
	if e == nil {
		return pipResource(models.EntityTypeEntity, uid, "", nil)
	}

	return pipResource(models.EntityTypeEntity, uid, e.StatusName(), e.Tags())
}

// submittedEntity is the resource of an entity that is not stored yet, carrying the submitted
// properties. A new entity is stored as concept whatever the body says.
func submittedEntity(in *oas.Entity) *models.Entity {
	return pipResource(models.EntityTypeEntity, models.EntityUID(in.Type, in.Id), models.StatusConcept.String(), in.Metadata.Tags)
}

// asConcept is the submitted entity with the status a new entity gets.
func asConcept(in *oas.Entity) *oas.Entity {
	out := *in
	out.Status = models.StatusConcept.String()

	return &out
}

// withStatus is the submitted entity with the status of the stored one it replaces.
func withStatus(in *oas.Entity, prev *models.Entity) *oas.Entity {
	out := *in
	out.Status = prev.StatusName()

	return &out
}
