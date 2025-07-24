package openfga_embedded

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/language/pkg/go/transformer"
	tuple2 "github.com/openfga/openfga/pkg/tuple"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Handle implements the EventSink interface.
func (c *controller) Handle(t models.EventType, key string) {
	switch t {
	case models.PolicyAdded, models.PolicyReplaced:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.OPENFGA.Language()) && !strings.EqualFold(language, models.OPENFGA.String()) {
			return
		}

		f, _, err := c.PAP().Read(language, id)
		if err != nil {
			c.Logger().Error("failed to get file", "controller", c.String(), "policy-id", id, "error", err)
			return
		}

		ext := filepath.Ext(id)
		store := strings.Replace(filepath.Base(id), ext, "", 1)

		switch ext {
		case ".mdl", ".model":
			c.addModel(store, f.Content())
		case ".rel", ".relations":
			c.addRelations(store, f.Content())
		}

	case models.PolicyRemoved:
		language, id := models.SplitPolicyKey(key)
		if !strings.EqualFold(language, models.OPENFGA.Language()) {
			return
		}

		ext := filepath.Ext(id)
		store := strings.Replace(filepath.Base(id), ext, "", 1)

		switch ext {
		case ".mdl", ".model":
			c.removeModel(store)
		case ".rel", ".relations":
			c.removeRelations(store)
		}

	default:
		// TODO: attributes, entities, relations
	}
}

func (c *controller) addModel(store string, f io.Reader) {
	d, _ := io.ReadAll(f)

	model, err := transformer.TransformDSLToProto(string(d))
	if err != nil {
		c.Logger().Error("failed to compile model", "controller", c.String(), "store", store, "error", err)
		return
	}

	dtl := c.findStore(store)
	if dtl == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	resp, err2 := c.engine.WriteAuthorizationModel(context.Background(), &openfgav1.WriteAuthorizationModelRequest{
		StoreId:         dtl.storeID,
		TypeDefinitions: model.GetTypeDefinitions(),
		Conditions:      model.GetConditions(),
		SchemaVersion:   model.GetSchemaVersion(),
	})
	if err2 != nil {
		c.Logger().Error("failed to add/replace model", "controller", c.String(), "store", store, "error", err2)
		return
	}

	dtl.authModelID = resp.GetAuthorizationModelId()
	c.Logger().Info("model added/replaced", "controller", c.String(), "store", store, "storeID", dtl.storeID, "authModelID", dtl.authModelID)
}

func (c *controller) removeModel(store string) {
	dtl, ok := c.stores[store]
	if !ok {
		return // nothing here.
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.engine.DeleteStore(context.Background(), &openfgav1.DeleteStoreRequest{StoreId: dtl.storeID})
	if err != nil {
		c.Logger().Error("failed to remove store", "controller", c.String(), "store", store, "storeID", dtl.storeID, "error", err)
		return
	}

	delete(c.stores, store)
	c.Logger().Info("store removed", "controller", c.String(), "store", store, "storeID", dtl.storeID)
}

func (c *controller) addRelations(store string, f io.Reader) {
	dtl := c.findStore(store)
	if dtl == nil {
		return
	}

	writes, deletes := c.buildRelationUpdates(store, f)
	if (writes == nil || len(writes.TupleKeys) == 0) && (deletes == nil || len(deletes.TupleKeys) == 0) {
		return
	}

	req := &openfgav1.WriteRequest{StoreId: dtl.storeID, AuthorizationModelId: dtl.authModelID}
	if writes != nil && len(writes.TupleKeys) > 0 {
		req.Writes = writes
	}
	if deletes != nil && len(deletes.TupleKeys) > 0 {
		req.Deletes = deletes
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.engine.Write(context.Background(), req)
	if err != nil {
		c.Logger().Error("failed to add/replace relations", "controller", c.String(), "store", store, "storeID", dtl.storeID, "error", err)
		return
	}

	// remember the new list of keys.
	clear(dtl.relations)
	if writes != nil {
		for _, t := range writes.TupleKeys {
			dtl.relations[keyFromTuple(t)] = struct{}{}
		}
	}

	c.Logger().Info("relations added/replaced", "controller", c.String(), "store", store, "storeID", dtl.storeID)
}

func (c *controller) removeRelations(store string) {
	dtl := c.findStore(store)
	if dtl == nil {
		return
	}

	// all relations will be removed.
	deletes := &openfgav1.WriteRequestDeletes{}
	for key := range dtl.relations {
		items := strings.Split(key, "|")
		t := tuple2.NewTupleKey(items[0], items[1], items[2])
		deletes.TupleKeys = append(deletes.TupleKeys, tuple2.TupleKeyToTupleKeyWithoutCondition(t))
	}

	if len(deletes.TupleKeys) == 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.engine.Write(context.Background(), &openfgav1.WriteRequest{
		StoreId:              dtl.storeID,
		Deletes:              deletes,
		AuthorizationModelId: dtl.authModelID,
	})
	if err != nil {
		c.Logger().Error("failed to remove relations", "controller", c.String(), "store", store, "storeID", dtl.storeID, "error", err)
		return
	}

	// clear the list of keys.
	clear(dtl.relations)
	c.Logger().Info("relations removed", "controller", c.String(), "store", store, "storeID", dtl.storeID)
}

func (c *controller) findStore(store string) *details {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	dtl, ok := c.stores[store]
	if ok {
		return dtl
	}

	s, err2 := c.engine.CreateStore(context.Background(), &openfgav1.CreateStoreRequest{Name: store})
	if err2 != nil {
		c.Logger().Error("failed to create store", "store", store, "error", err2)
		return nil
	}

	storeID := s.GetId()
	dtl = &details{store: store, storeID: storeID, relations: make(map[string]struct{})}
	c.stores[store] = dtl

	return dtl
}

func (c *controller) buildRelationUpdates(store string, f io.Reader) (*openfgav1.WriteRequestWrites, *openfgav1.WriteRequestDeletes) {
	list, err := readTuples(f)
	if err != nil {
		c.Logger().Error("failed to decode relations", "store", store, "error", err)
		return nil, nil
	}

	writes := &openfgav1.WriteRequestWrites{}
	deletes := &openfgav1.WriteRequestDeletes{}

	// determine the new relations = the writes.
	keys := make(map[string]*openfgav1.TupleKey, len(list))
	for _, t := range list {
		keys[keyFromTuple(t)] = t
		writes.TupleKeys = append(writes.TupleKeys, t)
	}

	// determine which old relations are no longer present = the deletes.
	if dtl, ok := c.stores[store]; ok {
		for key := range dtl.relations {
			if _, ok2 := keys[key]; !ok2 {
				items := strings.Split(key, "|")
				t := tuple2.NewTupleKey(items[0], items[1], items[2])
				deletes.TupleKeys = append(deletes.TupleKeys, tuple2.TupleKeyToTupleKeyWithoutCondition(t))
			}
		}
	}

	return writes, deletes
}

func keyFromTuple(t *openfgav1.TupleKey) string {
	return fmt.Sprintf("%s|%s|%s", t.User, t.Relation, t.Object)
}
