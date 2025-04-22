package openfga_embedded

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/openfga/pkg/tuple"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, parc *models.PARC) (*models.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	out := &models.Response{Allowed: false}

	fgaReq, msg := c.buildCheckRequest(parc)
	if msg != "" {
		out.Message = msg
		if debug {
			c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", uid, "message", msg)
		}
		return out, nil
	}

	started := time.Now()
	resp, err := c.engine.Check(context.Background(), fgaReq)
	duration := time.Since(started)

	if err != nil {
		out.Message = err.Error()
	} else {
		out.Allowed = resp.Allowed
		out.Message = resp.Resolution
	}

	if debug {
		if out.Allowed {
			c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", uid, "pdp elapsed", duration.String())
		} else {
			c.Logger().Warn("authorization failed", "controller", c.String(), "request-uid", uid, "message", out.Message, "pdp elapsed", duration.String())
		}
	}

	return out, nil
}

func (c *controller) buildCheckRequest(parc *models.PARC) (*openfgav1.CheckRequest, string) {
	parc = c.Map(parc)

	storeID := parc.Principal.Type()

	dtl, ok := c.stores[storeID]
	if !ok || dtl == nil {
		return nil, fmt.Sprintf("store not found; invalid principal type '%s'", storeID)
	}

	t := tupleKeyFromBasicTuple(basicTuple{
		Subject:   basicEntity{Type: parc.Principal.Type(), ID: parc.Principal.ID()},
		Predicate: parc.Action.ID(),
		Object:    basicEntity{Type: parc.Resource.Type(), ID: normalize(parc.Resource.ID())},
	})

	// TODO: add the context!
	return &openfgav1.CheckRequest{
		StoreId:              dtl.storeID,
		AuthorizationModelId: dtl.authModelID,
		TupleKey:             tuple.NewCheckRequestTupleKey(t.Object, t.Relation, t.User),
		Trace:                true,
		Consistency:          openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
	}, ""
}
