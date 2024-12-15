package openfga

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/openfga/pkg/tuple"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *components.Request) (*components.Response, error) {
	debug := c.Logger().Enabled(nil, slog.LevelDebug)
	out := &components.Response{Allowed: false}

	fgaReq, msg := c.buildCheckRequest(req)
	if msg != "" {
		out.Message = msg
		if debug {
			c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", req.UID, "message", out.Message)
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
			c.Logger().Debug("authorization granted", "controller", c.String(), "request-uid", req.UID, "pdp elapsed", duration.String())
		} else {
			c.Logger().Error("authorization failed", "controller", c.String(), "request-uid", req.UID, "message", out.Message, "pdp elapsed", duration.String())
		}
	}

	return out, nil
}

func (c *controller) buildCheckRequest(req *components.Request) (*openfgav1.CheckRequest, string) {
	a, uri := c.PIP().CollectAttributesFromRequest(req)
	if uri == "" && req.URL != nil {
		uri = req.URL.String()
	}

	p1, p2 := models.DeterminePrincipal(a)

	dtl, ok := c.stores[p1]
	if !ok || dtl == nil {
		return nil, fmt.Sprintf("store not found; invalid principal type '%s'", p1)
	}

	t := tupleKeyFromBasicTuple(basicTuple{
		Subject:   basicEntity{Type: p1, ID: p2},
		Predicate: "call",
		Object:    basicEntity{Type: models.EntityService, ID: normalize(uri)},
	})

	out := &openfgav1.CheckRequest{}
	out.StoreId = dtl.storeID
	out.AuthorizationModelId = dtl.authModelID
	out.TupleKey = tuple.NewCheckRequestTupleKey(t.Object, t.Relation, t.User)
	out.Trace = true
	out.Consistency = openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY
	return out, ""
}
