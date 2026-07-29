package settings

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
)

// SettingsPersister stores and retrieves manager ui settings.
type SettingsPersister interface {
	GetSettings(ctx context.Context) (*oas.Settings, error)
	UpdateSettings(ctx context.Context, user identity.Principal, in *oas.Settings) (*oas.Settings, error)
}

// Default settings, used by Defaults when nothing has been saved yet (see oas/settings/openapi.yaml).
const (
	DefaultHeaderTitle = "OpenFTV beheeromgeving"
	DefaultHeaderColor = "#F7E8E8"
	DefaultTitleColor  = "#000000"
)

// Defaults returns the built-in manager ui settings, used when nothing has been saved yet.
// Each call returns a fresh value, safe for the caller to mutate.
func Defaults() *oas.Settings {
	return &oas.Settings{
		HeaderTitle: DefaultHeaderTitle,
		HeaderColor: DefaultHeaderColor,
		TitleColor:  DefaultTitleColor,
	}
}
