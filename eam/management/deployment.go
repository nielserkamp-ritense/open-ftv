package management

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	kvstore "github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

// ErrBundle reports that a bundle could not be produced: an unknown bundle id, or a deployment
// refused because the last one is still running. It is only returned to a caller who was
// permitted the operation; any other failure is the store's and stays unexplained.
var ErrBundle = errors.New("bundle")

// DeploymentService administers deployments: every operation is one decision followed by one
// store call. Publishing a deployment is the `deploy` action (ADR 0006), decided on the
// collection because the deployment's version does not exist before the store generates it.
type DeploymentService struct {
	logger     *slog.Logger
	store      *pap.PAP
	manager    *bundles.Manager
	configs    []*bundles.Config // fixed at startup, so read once.
	authorizer authorization.Authorizer
}

// NewDeploymentService instantiates the deployment service. A nil authorizer permits
// everything, for deployments and tests that run without authorization.
func NewDeploymentService(logger *slog.Logger, store *pap.PAP, manager *bundles.Manager, authorizer authorization.Authorizer) *DeploymentService {
	return &DeploymentService{logger: logger, store: store, manager: manager, configs: manager.Bundles(), authorizer: authorizer}
}

// Configs returns the bundle configurations the manager serves.
func (s *DeploymentService) Configs(ctx context.Context, caller *authorization.RequestPrincipal) ([]*bundles.Config, error) {
	if err := s.authorize(ctx, caller, models.ActionRead, deploymentCollection()); err != nil {
		return nil, err
	}

	return s.configs, nil
}

// Describe permits reading the vocabulary of deployments (statuses, compression types), which
// is decided as a read of the collection.
func (s *DeploymentService) Describe(ctx context.Context, caller *authorization.RequestPrincipal) error {
	return s.authorize(ctx, caller, models.ActionRead, deploymentCollection())
}

// DecideDeploy asks whether caller may publish, without publishing: it lets a handler answer a
// malformed deploy request in decision order, the denial before the parse error.
func (s *DeploymentService) DecideDeploy(ctx context.Context, caller *authorization.RequestPrincipal) error {
	return s.authorize(ctx, caller, models.ActionDeploy, deploymentCollection())
}

// List returns every deployment the store holds.
func (s *DeploymentService) List(ctx context.Context, caller *authorization.RequestPrincipal) ([]*bundles.Deployment, error) {
	if err := s.authorize(ctx, caller, models.ActionRead, deploymentCollection()); err != nil {
		return nil, err
	}

	return s.store.ListDeployments()
}

// Get returns one deployment by version.
func (s *DeploymentService) Get(ctx context.Context, caller *authorization.RequestPrincipal, version uint64) (*bundles.Deployment, error) {
	d, err := s.load(ctx, caller, version)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, caller, models.ActionRead, deploymentResource(version, d)); err != nil {
		return nil, err
	}

	if d == nil {
		return nil, ErrNotFound
	}

	return d, nil
}

// Last returns the most recent deployment, or ErrNotFound when none was ever made. The
// decision is on the deployment the store returned, and a failed or empty read is decided on
// the collection, so the store's outcome reaches permitted callers only.
func (s *DeploymentService) Last(ctx context.Context, caller *authorization.RequestPrincipal) (*bundles.Deployment, error) {
	last, err := s.store.LastDeployment()
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		last, err = nil, nil
	}

	if err != nil || last == nil || last.Version() <= 0 {
		if derr := s.authorize(ctx, caller, models.ActionRead, deploymentCollection()); derr != nil {
			return nil, derr
		}

		if err != nil {
			return nil, err
		}

		return nil, ErrNotFound
	}

	if err := s.authorize(ctx, caller, models.ActionRead, deploymentResource(last.Version(), last)); err != nil {
		return nil, err
	}

	return last, nil
}

// Bundle returns the bundle with the given id from the last deployment, which is how a PDP
// fetches what was published to it.
func (s *DeploymentService) Bundle(ctx context.Context, caller *authorization.RequestPrincipal, id string) (*bundles.Bundle, error) {
	last, err := s.Last(ctx, caller)
	if err != nil {
		return nil, err
	}

	b, err := s.manager.Get(last, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBundle, err)
	}

	return b, nil
}

// Deploy publishes a new deployment: a bundle run over the current policies and data,
// attributed to the caller.
func (s *DeploymentService) Deploy(ctx context.Context, caller *authorization.RequestPrincipal, title, description string) (*bundles.Deployment, error) {
	if err := s.authorize(ctx, caller, models.ActionDeploy, deploymentCollection()); err != nil {
		return nil, err
	}

	d, err := s.store.NewDeployment(title, description, s.manager, caller.Attribution())
	if errors.Is(err, bundles.ErrDeploymentActive) {
		return nil, fmt.Errorf("%w: %w", ErrBundle, err)
	}

	return d, err
}

// load reads the deployment an operation acts on; a missing one is nil. A read that fails is
// still decided, on the id-only resource, and the decision is reported first: a caller the
// PDP denies gets the denial and never the store's error.
func (s *DeploymentService) load(ctx context.Context, caller *authorization.RequestPrincipal, version uint64) (*bundles.Deployment, error) {
	d, err := s.store.ReadDeployment(version)
	if err == nil || errors.Is(err, kvstore.ErrKeyNotFound) {
		return d, nil
	}

	if derr := s.authorize(ctx, caller, models.ActionRead, deploymentResource(version, nil)); derr != nil {
		return nil, derr
	}

	return nil, err
}

// authorize asks the PDP; without an authorizer every operation is permitted.
func (s *DeploymentService) authorize(ctx context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	if s.authorizer == nil {
		return nil
	}

	return s.authorizer.Authorize(ctx, caller, action, resource)
}

// deploymentCollection is the resource of a collection read, and of publishing a deployment.
func deploymentCollection() *models.Entity {
	return models.NewEntity(models.EntityTypeDeployment, models.ResourceCollection, models.NewAttributeSet())
}

// deploymentResource is the resource of one deployment, identified by its version and carrying
// its status. A nil object (not stored) contributes its id only.
func deploymentResource(version uint64, d *bundles.Deployment) *models.Entity {
	attrs := models.NewAttributeSet()

	if d != nil {
		attrs.AddAttributeKV(models.AttrStatus, d.Status().Status())
	}

	return models.NewEntity(models.EntityTypeDeployment, strconv.FormatUint(version, 10), attrs)
}
