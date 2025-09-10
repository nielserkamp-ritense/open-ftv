package bundles

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/goccy/go-json"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// NewKeyValueDB instantiates a new bundle deployment persistence manager with a KV store as backend.
func NewKeyValueDB(client store.Store, basePath string) *KeyValueDB {
	return &KeyValueDB{client: client, basePath: convert.ForceSuffix(basePath, pathSeparator)}
}

// KeyValueDB contains the details to manage bundle deployments using a KV store as backend.
type KeyValueDB struct {
	client   store.Store
	basePath string
	mutex    sync.RWMutex
}

// Generate creates a new deployment in the store.
//
// If the last deployment is still active, this function will return an error.
func (s *KeyValueDB) Generate(ctx context.Context, description, user string) (*Deployment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion(ctx)
	if err != nil && err.Error() != "key not found in store" {
		return nil, err
	}

	if err == nil {
		key := s.makeKey("deploy", strconv.FormatUint(version, 10))
		deploy, err3 := s.readDeployment(ctx, key)
		if err3 != nil {
			return nil, err3
		}

		if status := deploy.Status(); status < Failed {
			return nil, fmt.Errorf("deployment %d is still active [%s]", version, status.String())
		}
	}

	version++
	deploy := NewDeployment(version, description, user)

	if err = s.putVersion(ctx, version); err != nil {
		return nil, err
	}
	if err = s.putDeployment(ctx, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// Advance advances the status of the last deployment in the store.
//
// If the current status is *Sending*, the status will be set to *Completed*.
// To set the status to *Failed*, use the Fail() function.
//
// If the last deployment is no longer active, this function will return an error.
func (s *KeyValueDB) Advance() (*Deployment, error) {
	ctx := context.Background()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	key := s.makeKey("deploy", strconv.FormatUint(version, 10))
	deploy, err3 := s.readDeployment(ctx, key)
	if err3 != nil {
		return nil, err3
	}

	status := deploy.Status()
	switch status {
	case Failed, Completed:
		return nil, fmt.Errorf("deployment %d no longer active [%s]", version, status.String())
	case Sending:
		deploy.Completed()
	default:
		deploy.NextStatus()
	}

	if err = s.putDeployment(ctx, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// Fail sets the status of the last deployment in the store to *Failed*.
//
// If the last deployment is no longer active, this function will return an error.
func (s *KeyValueDB) Fail(msg string) (*Deployment, error) {
	ctx := context.Background()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	key := s.makeKey("deploy", strconv.FormatUint(version, 10))
	deploy, err3 := s.readDeployment(ctx, key)
	if err3 != nil {
		return nil, err3
	}

	status := deploy.Status()
	switch status {
	case Failed, Completed:
		return nil, fmt.Errorf("deployment %d no longer active [%s]", version, status.String())
	default:
		deploy.Failed(msg)
	}

	if err = s.putDeployment(ctx, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// LastDeployment retrieves the last deployment from the store.
func (s *KeyValueDB) LastDeployment(ctx context.Context) (*Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	version, err := s.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	key := s.makeKey("deploy", strconv.FormatUint(version, 10))
	return s.readDeployment(ctx, key)
}

// ReadDeployment retrieves a deployment from the store.
func (s *KeyValueDB) ReadDeployment(ctx context.Context, version uint64) (*Deployment, error) {
	key := s.makeKey("deploy", strconv.FormatUint(version, 10))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.readDeployment(ctx, key)
}

func (s *KeyValueDB) readDeployment(ctx context.Context, key string) (*Deployment, error) {
	kv, err := s.client.Get(ctx, key, readOptions)
	if err != nil {
		return nil, err
	}

	deploy, err2 := s.unmarshalDeployment(kv)
	if err2 != nil {
		return nil, err2
	}
	return deploy, nil
}

// ListDeployments retrieves all deployments from the store.
func (s *KeyValueDB) ListDeployments(ctx context.Context) ([]*Deployment, error) {
	key := s.makeKey("deploy", "")

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(ctx, key, readOptions)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read deployments: %w", err)
	}

	out := make([]*Deployment, 0, len(list))
	for _, kv := range list {
		p := new(Deployment)
		if err = json.Unmarshal(kv.Value, p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal deployments: %w", err)
		}
		out = append(out, p)
	}

	return out, nil
}

// CreateBundleAudit is not implemented for KV stores.
func (s *KeyValueDB) CreateBundleAudit(context.Context, uint64, *Config, *Bundle) error {
	return errors.New("not implemented")
}

func (s *KeyValueDB) getVersion(ctx context.Context) (uint64, error) {
	key := s.makeKey("version", "last")
	kv, err := s.client.Get(ctx, key, readOptions)
	if err != nil {
		return 0, err
	}
	return s.unmarshalVersion(kv)
}

func (s *KeyValueDB) putVersion(ctx context.Context, version uint64) error {
	key := s.makeKey("version", "last")
	b, _ := json.Marshal(version)
	return s.client.Put(ctx, key, b, writeOptions)
}

func (s *KeyValueDB) putDeployment(ctx context.Context, d *Deployment) error {
	key := s.makeKey("deploy", strconv.FormatUint(d.Version(), 10))
	b, _ := json.Marshal(d)
	return s.client.Put(ctx, key, b, writeOptions)
}

func (s *KeyValueDB) makeKey(prefix, id string) string {
	return fmt.Sprintf("%s$$$%s$$$%s%s", s.basePath, prefix, pathSeparator, id)
}

func (s *KeyValueDB) mustMarshal(p *KeyValueDB) []byte {
	b, _ := json.Marshal(p)
	return b
}

func (s *KeyValueDB) unmarshalVersion(kv *store.KVPair) (uint64, error) {
	var v uint64
	if err := json.Unmarshal(kv.Value, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (s *KeyValueDB) unmarshalDeployment(kv *store.KVPair) (*Deployment, error) {
	deploy := &Deployment{}
	if err := deploy.UnmarshalJSON(kv.Value); err != nil {
		return nil, err
	}
	return deploy, nil
}

func (s *KeyValueDB) bugFix(in string) string {
	// the Valkeyrie/etcdv3 implementation sometimes removes a leading slash character from the key.
	if _, ok := s.client.(*etcdv3.Store); ok {
		return "/" + in
	}
	return in
}

var (
	pathSeparator = "/"
	readOptions   = &store.ReadOptions{}
	writeOptions  = &store.WriteOptions{}
)
