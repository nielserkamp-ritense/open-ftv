package openfga

import (
	"context"
	"io"
	"path/filepath"

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
			c.Logger().Error("failed to get policy", "controller", c.String(), "key", key, "error", err)
			return
		}

		switch filepath.Ext(key) {
		case ".mdl", ".model":
			c.addModel(key, f)
		case ".rel", ".relations":
			c.addRelations(key, f)
		}

	case models.PolicyRemoved:
		switch filepath.Ext(key) {
		case ".mdl", ".model":
			c.removeModel(key)
		case ".rel", ".relations":
			c.removeRelations(key)
		}

	}
}

func (c *controller) addModel(key string, f io.Reader) {
	store := filepath.Base(key)
	d, _ := io.ReadAll(f)

	model, err := transformer.TransformDSLToProto(string(d))
	if err != nil {
		c.Logger().Error("failed to compile model", "controller", c.String(), "model-key", key, "error", err)
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	storeID := c.stores[store]
	if storeID == "" {
		s, err2 := c.engine.CreateStore(
			context.Background(),
			&openfgav1.CreateStoreRequest{Name: store},
		)
		if err2 != nil {
			c.Logger().Error("Failed to create store", "error", err2)
			return
		}

		storeID = s.GetId()
		c.stores[store] = storeID
	}

	resp, err2 := c.engine.WriteAuthorizationModel(context.Background(), &openfgav1.WriteAuthorizationModelRequest{
		StoreId:         storeID,
		TypeDefinitions: model.GetTypeDefinitions(),
		Conditions:      model.GetConditions(),
		SchemaVersion:   model.GetSchemaVersion(),
	})
	if err2 != nil {
		c.Logger().Error("failed to add model", "controller", c.String(), "model-key", key, "error", err2)
		return
	}

	policyID := resp.GetAuthorizationModelId()
	c.models[storeID] = policyID

	c.Logger().Info("model added/replaced", "controller", c.String(), "model-key", key, "storeID", storeID, "modelID", policyID)
}

func (c *controller) removeModel(key string) {
	store := filepath.Base(key)
	storeID := c.stores[store]
	if storeID == "" {
		return // nothing here.
	}

	_, err := c.engine.DeleteStore(context.Background(), &openfgav1.DeleteStoreRequest{StoreId: storeID})
	if err != nil {
		c.Logger().Error("failed to remove store", "controller", c.String(), "model-key", key, "storeID", storeID, "error", err)
		return
	}

	c.Logger().Info("store removed", "controller", c.String(), "model-key", key, "storeID", storeID)
}

func (c *controller) addRelations(key string, f io.Reader) {
	// TODO: ...
}

func (c *controller) removeRelations(key string) {
	// TODO: ...
}
