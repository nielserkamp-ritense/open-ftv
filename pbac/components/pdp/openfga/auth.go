package openfga

import (
	"context"
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	tuple2 "github.com/openfga/openfga/pkg/tuple"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(req *components.Request) (*components.Response, error) {
	out := &components.Response{Allowed: false}

	store := req.Principal.Type()
	dtl, ok := c.stores[store]
	if !ok {
		out.Message = fmt.Sprintf("store not found; invalid principal type '%s'", req.Principal.Type())
		return out, nil
	}

	resp, err := c.engine.Check(context.Background(), c.buildCheckRequest(req, dtl))
	if err != nil {
		out.Message = err.Error()
	} else {
		out.Allowed = resp.Allowed
		out.Message = resp.Resolution
	}

	return out, nil
}

func (c *controller) buildCheckRequest(req *components.Request, dtl *details) *openfgav1.CheckRequest {
	t := tupleKeyFromTuple(tuple{
		Subject:   tupleEntity{Type: req.Principal.Type(), ID: req.Principal.ID()},
		Predicate: "call",
		Object:    tupleEntity{Type: req.Resource.Type(), ID: req.Resource.ID()},
	})

	out := &openfgav1.CheckRequest{
		StoreId:              dtl.storeID,
		TupleKey:             tuple2.NewCheckRequestTupleKey(t.User, t.Relation, t.Object),
		AuthorizationModelId: dtl.authModelID,
		Trace:                true,
		Consistency:          openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
	}

	return out
}
