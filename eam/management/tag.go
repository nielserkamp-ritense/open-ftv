package management

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// TagStore is what the tag service needs from the PAP: the tag operations, without the rest.
// *pap.PAP satisfies it.
type TagStore interface {
	ListTags() (oas.Tags, error)
	ReadTag(id string) (*oas.Tag, uint64, error)
	CreateTag(in *oas.Tag, user identity.Principal) (*oas.Tag, error)
	UpdateTag(prev *oas.Tag, lastIndex uint64, in *oas.Tag, user identity.Principal) (*oas.Tag, error)
	DeleteTag(prev *oas.Tag, lastIndex uint64, user identity.Principal) (*oas.Tag, error)
}

// TagService administers tags, the labels that name a beslispunt (see CONTEXT.md): every
// operation is one decision followed by one store call.
type TagService struct {
	logger     *slog.Logger
	store      TagStore
	authorizer authorization.Authorizer
}

// NewTagService instantiates the tag service. A nil authorizer permits everything, for
// deployments and tests that run without authorization.
func NewTagService(logger *slog.Logger, store TagStore, authorizer authorization.Authorizer) *TagService {
	return &TagService{logger: logger, store: store, authorizer: authorizer}
}

// List returns every tag the store holds.
func (s *TagService) List(ctx context.Context, caller *authorization.RequestPrincipal) (oas.Tags, error) {
	if err := s.authorize(ctx, caller, models.ActionRead, tagCollection()); err != nil {
		return nil, err
	}

	return s.store.ListTags()
}

// Get returns one tag.
func (s *TagService) Get(ctx context.Context, caller *authorization.RequestPrincipal, id string) (*oas.Tag, error) {
	prev, _, err := s.load(ctx, caller, models.ActionRead, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, tagResource(id)); err != nil {
		return nil, err
	}

	if prev == nil {
		return nil, ErrNotFound
	}

	return prev, nil
}

// Create adds a new tag. With forceUpsert an existing tag is updated instead, which is then
// decided as an update; without it an existing id is ErrExists.
func (s *TagService) Create(ctx context.Context, caller *authorization.RequestPrincipal, t *oas.Tag, forceUpsert bool) (*oas.Tag, error) {
	prev, lastIndex, err := s.load(ctx, caller, models.ActionCreate, t.Id)
	if err != nil {
		return nil, err
	}

	switch {
	case prev == nil:
		if err := s.authorize(ctx, caller, models.ActionCreate, tagResource(t.Id)); err != nil {
			return nil, err
		}

		return s.store.CreateTag(t, caller.Attribution())

	case !forceUpsert:
		if err := s.authorize(ctx, caller, models.ActionCreate, tagResource(t.Id)); err != nil {
			return nil, err
		}

		return nil, ErrExists

	default:
		if err := s.authorize(ctx, caller, models.ActionUpdate, tagResource(t.Id)); err != nil {
			return nil, err
		}

		return s.store.UpdateTag(prev, lastIndex, t, caller.Attribution())
	}
}

// Update replaces an existing tag. With forceUpsert a missing tag is created instead, which
// is then decided as a create; without it a missing id is ErrNotFound.
func (s *TagService) Update(ctx context.Context, caller *authorization.RequestPrincipal, t *oas.Tag, forceUpsert bool) (*oas.Tag, error) {
	prev, lastIndex, err := s.load(ctx, caller, models.ActionUpdate, t.Id)
	if err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		if err := s.authorize(ctx, caller, models.ActionUpdate, tagResource(t.Id)); err != nil {
			return nil, err
		}

		return s.store.UpdateTag(prev, lastIndex, t, caller.Attribution())

	case !forceUpsert:
		if err := s.authorize(ctx, caller, models.ActionUpdate, tagResource(t.Id)); err != nil {
			return nil, err
		}

		return nil, ErrNotFound

	default:
		if err := s.authorize(ctx, caller, models.ActionCreate, tagResource(t.Id)); err != nil {
			return nil, err
		}

		return s.store.CreateTag(t, caller.Attribution())
	}
}

// Delete removes a tag. With ignoreMissing a missing id yields an empty tag carrying the id,
// so the call is idempotent; without it a missing id is ErrNotFound.
func (s *TagService) Delete(ctx context.Context, caller *authorization.RequestPrincipal, id string, ignoreMissing bool) (*oas.Tag, error) {
	prev, lastIndex, err := s.load(ctx, caller, models.ActionDelete, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionDelete, tagResource(id)); err != nil {
		return nil, err
	}

	switch {
	case prev != nil:
		return s.store.DeleteTag(prev, lastIndex, caller.Attribution())
	case ignoreMissing:
		return &oas.Tag{Id: id}, nil
	default:
		return nil, ErrNotFound
	}
}

// load reads the tag an operation acts on. A read that fails is still decided, and the
// decision is reported first, so the store's state and its error text reach permitted
// callers only.
func (s *TagService) load(ctx context.Context, caller *authorization.RequestPrincipal, action, id string) (*oas.Tag, uint64, error) {
	prev, lastIndex, err := s.store.ReadTag(id)
	if err == nil {
		return prev, lastIndex, nil
	}

	if derr := s.authorize(ctx, caller, action, tagResource(id)); derr != nil {
		return nil, 0, derr
	}

	return nil, 0, err
}

// authorize asks the PDP; without an authorizer every operation is permitted.
func (s *TagService) authorize(ctx context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	if s.authorizer == nil {
		return nil
	}

	return s.authorizer.Authorize(ctx, caller, action, resource)
}

// tagCollection is the resource of a collection read.
func tagCollection() *models.Entity {
	return models.NewEntity(models.EntityTypeTag, models.ResourceCollection, models.NewAttributeSet())
}

// tagResource is the resource of one tag. A tag has no properties a policy decides on beyond
// its id, so a stored and a missing tag look the same to the PDP.
func tagResource(id string) *models.Entity {
	return models.NewEntity(models.EntityTypeTag, id, models.NewAttributeSet())
}
