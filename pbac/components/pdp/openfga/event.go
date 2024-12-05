package openfga

import (
	"context"
	"io"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/language/pkg/go/transformer"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(t models.EventType, key string) {
	switch t {
	case models.PolicyAdded, models.PolicyReplaced:
		f, err := c.PAP().Get(key)
		if err != nil {
			c.Logger().Error("failed to get policy", "controller", c.String(), "policy-key", key, "error", err)
			return
		}

		d, _ := io.ReadAll(f)

		var model *openfgav1.AuthorizationModel
		model, err = transformer.TransformDSLToProto(string(d))
		if err != nil {
			c.Logger().Error("failed to compile policy", "controller", c.String(), "policy-key", key, "error", err)
			return
		}

		_, err = c.pdp.WriteAuthorizationModel(context.Background(), &openfgav1.WriteAuthorizationModelRequest{
			StoreId:         c.storeID,
			TypeDefinitions: model.GetTypeDefinitions(),
			Conditions:      model.GetConditions(),
			SchemaVersion:   model.GetSchemaVersion(),
		})
		if err != nil {
			c.Logger().Error("failed to add policy", "controller", c.String(), "policy-key", key, "error", err)
			return
		}

		c.Logger().Info("policy added/replaced", "controller", c.String(), "policy-key", key)

	case models.PolicyRemoved:
		// OpenFGA does not support removal of policies.
		// c.ds.Remove(cedar.PolicyID(key))
		// c.Logger().Info("policy removed", "controller", c.String(), "policy-key", key)
	}
}
