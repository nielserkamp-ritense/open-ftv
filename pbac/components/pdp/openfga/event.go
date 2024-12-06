package openfga

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/language/pkg/go/transformer"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(t models.EventType, key string) {
	ext := filepath.Ext(key)
	store := strings.Replace(filepath.Base(key), ext, "", 1)

	switch t {
	case models.PolicyAdded, models.PolicyReplaced:
		f, err := c.PAP().Get(key)
		if err != nil {
			c.Logger().Error("failed to get policy", "controller", c.String(), "key", key, "error", err)
			return
		}

		switch ext {
		case ".mdl", ".model":
			c.addModel(store, f)
		case ".rel", ".relations":
			c.addRelations(store, f)
		}

	case models.PolicyRemoved:
		switch ext {
		case ".mdl", ".model":
			c.removeModel(store)
		case ".rel", ".relations":
			c.removeRelations(store)
		}

	}
}

func (c *controller) addModel(store string, f io.Reader) {
	d, _ := io.ReadAll(f)

	model, err := transformer.TransformDSLToProto(string(d))
	if err != nil {
		c.Logger().Error("failed to compile model", "controller", c.String(), "store", store, "error", err)
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
			c.Logger().Error("Failed to create store", "store", store, "error", err2)
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
		c.Logger().Error("failed to add model", "controller", c.String(), "store", store, "error", err2)
		return
	}

	authID := resp.GetAuthorizationModelId()
	c.models[storeID] = authID

	c.Logger().Info("model added/replaced", "controller", c.String(), "store", store, "storeID", storeID, "authID", authID)
}

func (c *controller) removeModel(store string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	storeID := c.stores[store]
	if storeID == "" {
		return // nothing here.
	}

	_, err := c.engine.DeleteStore(context.Background(), &openfgav1.DeleteStoreRequest{StoreId: storeID})
	if err != nil {
		c.Logger().Error("failed to remove store", "controller", c.String(), "store", store, "storeID", storeID, "error", err)
		return
	}

	c.Logger().Info("store removed", "controller", c.String(), "store", store, "storeID", storeID)
}

func (c *controller) addRelations(store string, f io.Reader) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	storeID, ok := c.stores[store]
	if !ok {
		c.Logger().Error("failed to find store", "controller", c.String(), "store", store)
		return
	}

	authID, ok2 := c.models[storeID]
	if !ok2 {
		c.Logger().Error("failed to find authorization model", "controller", c.String(), "store", store, "storeID", storeID)
		return
	}

	writes, deletes := c.buildRelationUpdates(store, f)

	_, err := c.engine.Write(context.Background(), &openfgav1.WriteRequest{
		StoreId:              storeID,
		Writes:               writes,
		Deletes:              deletes,
		AuthorizationModelId: authID,
	})
	if err != nil {
		c.Logger().Error("failed to maintain relations", "controller", c.String(), "store", store, "storeID", storeID, "error", err)
	}
}

func (c *controller) removeRelations(key string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// TODO: ...

}

func (c *controller) buildRelationUpdates(store string, f io.Reader) (*openfgav1.WriteRequestWrites, *openfgav1.WriteRequestDeletes) {

	return nil, nil
}
