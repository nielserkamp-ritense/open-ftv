package openfga_embedded

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/openfga/pkg/tuple"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(uid string, parc *models.PARC) (*models.Response, error) {
	logger := c.Logger.With("controller", c.String(), "request-uid", uid)

	debug := c.Logger.Enabled(nil, slog.LevelDebug)
	if debug {
		logger.Debug("authorization request")
	}

	out := &models.Response{Allowed: false}

	fgaReq, err := c.buildCheckRequest(parc)
	if err != nil {
		out.Message = err.Error()
		if debug {
			logger.Error("authorization failed", "error", err)
		}
		return nil, err
	}

	c.AuthMutex.RLock()

	started := time.Now()
	resp, err2 := c.engine.Check(context.Background(), fgaReq)
	duration := time.Since(started)

	c.AuthMutex.RUnlock()

	switch {
	case err2 != nil:
		out.Message = err2.Error()
	case resp == nil:
		out.Message = "authorization failed; empty response from engine"
	default:
		out.Allowed = resp.Allowed
		out.Message = resp.Resolution
	}

	if debug {
		if out.Allowed {
			logger.Debug("authorization granted", "pdp elapsed", duration.String())
		} else {
			logger.Warn("authorization failed", "message", out.Message, "pdp elapsed", duration.String())
		}
	}

	return out, nil
}

// Batch implements the Controller interface.
func (c *controller) Batch(uid string, req *models.Batch) ([]models.Response, error) {
	logger := c.Logger.With("controller", c.String(), "request-uid", uid)

	debug := c.Logger.Enabled(nil, slog.LevelDebug)
	if debug {
		logger.Debug("batch request", "semantics", req.Semantics.String())
	}

	dtl, err := c.getStore(&req.Items[0])
	if err != nil {
		return nil, err
	}

	fgaReq := &openfgav1.BatchCheckRequest{
		StoreId:              dtl.storeID,
		AuthorizationModelId: dtl.authModelID,
		Checks:               make([]*openfgav1.BatchCheckItem, 0, len(req.Items)),
		Consistency:          openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
	}

	for i := range req.Items {
		r1 := &req.Items[i]
		r2, err2 := c.buildCheckRequest(r1)
		if err2 != nil {
			return nil, fmt.Errorf("failed to build check request for item #%d: %w", i+1, err2)
		}
		fgaReq.Checks = append(fgaReq.Checks, &openfgav1.BatchCheckItem{TupleKey: r2.TupleKey, CorrelationId: strconv.Itoa(i)})
	}

	c.AuthMutex.RLock()

	started := time.Now()
	resp, err2 := c.engine.BatchCheck(context.Background(), fgaReq)
	duration := time.Since(started)

	c.AuthMutex.RUnlock()

	if err2 != nil || resp == nil {
		logger.Error("batch failed", "err", err2, "pdp elapsed", duration.String())
		return nil, fmt.Errorf("failed to check batch: %w", err2)
	}

	out := make([]models.Response, 0, len(req.Items))
	var allowed int
	for i := range req.Items {
		decision := resp.Result[strconv.Itoa(i)]
		if decision.GetAllowed() {
			allowed++
			out = append(out, models.Response{Allowed: true})
		} else {
			var msg string
			if e := decision.GetError(); e != nil {
				msg = e.GetMessage()
			}
			out = append(out, models.Response{Allowed: false, Message: msg})
		}
	}

	logger.Error("batch succeeded", "allowed", allowed, "not allowed", len(req.Items)-allowed, "pdp elapsed", duration.String())
	return out, nil
}

func (c *controller) buildCheckRequest(parc *models.PARC) (*openfgav1.CheckRequest, error) {
	parc = c.Map(parc)

	dtl, err := c.getStore(parc)
	if err != nil {
		return nil, err
	}

	t := tupleKeyFromBasicTuple(basicTuple{
		Subject:   basicEntity{Type: parc.Principal.Type(), ID: parc.Principal.ID()},
		Predicate: parc.Action.ID(),
		Object:    basicEntity{Type: parc.Resource.Type(), ID: normalize(parc.Resource.ID())},
	})

	// TODO: add the context
	// TODO: add any contextual attributes and/or entities from the PIP

	return &openfgav1.CheckRequest{
		StoreId:              dtl.storeID,
		AuthorizationModelId: dtl.authModelID,
		TupleKey:             tuple.NewCheckRequestTupleKey(t.Object, t.Relation, t.User),
		Trace:                true,
		Consistency:          openfgav1.ConsistencyPreference_HIGHER_CONSISTENCY,
	}, nil
}

func (c *controller) getStore(parc *models.PARC) (*details, error) {
	// temporary solution until we have better use-cases and better understand how to model requests
	// and which request-attribute to use to pick the correct store.
	storeID := parc.Principal.Type()

	dtl, ok := c.stores[storeID]
	if !ok || dtl == nil {
		return nil, fmt.Errorf("store not found; invalid principal type '%s'", storeID)
	}

	return dtl, nil
}
