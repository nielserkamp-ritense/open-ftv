// Package selfauthz builds the embedded PDP that authorizes an app's own HTTP API.
package selfauthz

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	cedar_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	opa_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	openfga_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// NewController builds the embedded PDP that authorizes an app's own HTTP API for language,
// enforcing ap backed by ip, and using cerbos when language is CERBOS. Extra opts are applied
// last so they can override any default, e.g. a caller-specific pdp.WithPEP or pdp.WithADL.
func NewController(
	ctx context.Context, logger *slog.Logger, language models.Language, ip *pip.PIP, ap *pap.PAP,
	cerbos *cerbos_api.Config, opts ...pdp.Option,
) (pdp.Controller, error) {
	options := append([]pdp.Option{
		pdp.WithContext(ctx),
		pdp.WithLogger(logger),
		pdp.WithPEP(pep.New(ctx, logger)),
		pdp.WithPIP(ip),
		pdp.WithPAP(ap),
	}, opts...)

	switch language {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), nil
	case models.REGO:
		return opa_embedded.NewController(options...), nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), nil
	case models.CERBOS:
		return cerbos_api.NewController(*cerbos, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", language)
	}
}
