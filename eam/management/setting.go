package management

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/settings"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
)

// SettingUI is the id of the manager's UI settings: one record, upserted in place, so it is
// read and updated but never created or deleted (SCRUM-144, decided 2026-09-29). A later
// group of settings gets its own id; `resource is setting` covers them all.
const SettingUI = "ui"

// SettingService administers the manager's settings: every operation is one decision
// followed by one store call.
type SettingService struct {
	store      settings.SettingsPersister
	authorizer authorization.Authorizer
}

// NewSettingService instantiates the setting service. A nil authorizer permits everything,
// for deployments and tests that run without authorization.
func NewSettingService(store settings.SettingsPersister, authorizer authorization.Authorizer) *SettingService {
	return &SettingService{store: store, authorizer: authorizer}
}

// Get returns the UI settings, or the defaults when nothing has been saved yet.
func (s *SettingService) Get(ctx context.Context, caller *authorization.RequestPrincipal) (*oas.Settings, error) {
	if err := s.authorize(ctx, caller, models.ActionRead, settingResource(SettingUI)); err != nil {
		return nil, err
	}

	return s.store.GetSettings(ctx)
}

// Update replaces the UI settings. The record always exists to the caller (the defaults stand
// in for an unsaved one), so this is an update and never a create.
func (s *SettingService) Update(ctx context.Context, caller *authorization.RequestPrincipal, in *oas.Settings) (*oas.Settings, error) {
	if err := s.authorize(ctx, caller, models.ActionUpdate, settingResource(SettingUI)); err != nil {
		return nil, err
	}

	return s.store.UpdateSettings(ctx, caller.Attribution(), in)
}

// authorize asks the PDP; without an authorizer every operation is permitted.
func (s *SettingService) authorize(ctx context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	return decide(ctx, s.authorizer, caller, action, resource)
}

// settingResource is the resource of one settings record. It carries no properties.
func settingResource(id string) *models.Entity {
	return models.NewEntity(models.EntityTypeSetting, id, models.NewAttributeSet())
}
