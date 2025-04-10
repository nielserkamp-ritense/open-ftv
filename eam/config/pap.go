package config

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
)

// PAP contains the configuration variables for a generic PAP.
type PAP struct {
	Language     string `yaml:"policies.language,omitempty" env:"POLICIES_LANGUAGE" flag:"policies-language,language" default:"CEDAR" desc:"Language used for policy files"`
	Store        string `yaml:"policies.store.path,omitempty" env:"POLICIES_STORE" flag:"policies-store" desc:"Path where policy files are stored"`
	StoreRecurse bool   `yaml:"policies.store.recurse,omitempty" env:"POLICIES_STORE_RECURSE" flag:"policies-store-recurse" desc:"Search policy file storage recursively"`
}

// NewPAP instantiates a new PAP using the given configuration.
func (p *PAP) NewPAP(ctx context.Context, logger *slog.Logger) (pap.PAP, error) {
	return pap.New(ctx, logger, pap.WithLanguage(p.Language), pap.WithFileStore(p.Store, p.StoreRecurse)), nil
}
