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

// NewPersistence instantiates a new bundle deployment persistence manager.
func NewPersistence(ctx context.Context, client store.Store, basePath string) *Persistence {
	return &Persistence{ctx: ctx, client: client, basePath: convert.ForceSuffix(basePath, pathSeparator)}
}

// Persistence contains the details to manage bundle deployment persistence.
type Persistence struct {
	ctx      context.Context
	client   store.Store
	basePath string
	mutex    sync.RWMutex
}

// Generate creates a new deployment in the store.
//
// If the last deployment is still active, this function will return an error.
func (s *Persistence) Generate(description string) (*Deployment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion()
	if err != nil && err.Error() != "key not found in store" {
		return nil, err
	}

	if err == nil {
		key := s.makeKey("deploy", strconv.FormatUint(version, 10))
		deploy, err3 := s.readDeployment(key)
		if err3 != nil {
			return nil, err3
		}

		if status := deploy.Status(); status < Failed {
			return nil, fmt.Errorf("deployment %d is still active [%s]", version, status.String())
		}
	}

	version++
	deploy := NewDeployment(version, description)

	if err = s.putVersion(version); err != nil {
		return nil, err
	}
	if err = s.putDeployment(deploy); err != nil {
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
func (s *Persistence) Advance() (*Deployment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion()
	if err != nil {
		return nil, err
	}

	key := s.makeKey("deploy", strconv.FormatUint(version, 10))
	deploy, err3 := s.readDeployment(key)
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

	if err = s.putDeployment(deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// Fail sets the status of the last deployment in the store to *Failed*.
//
// If the last deployment is no longer active, this function will return an error.
func (s *Persistence) Fail(msg string) (*Deployment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion()
	if err != nil {
		return nil, err
	}

	key := s.makeKey("deploy", strconv.FormatUint(version, 10))
	deploy, err3 := s.readDeployment(key)
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

	if err = s.putDeployment(deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// LastDeployment retrieves the last deployment from the store.
func (s *Persistence) LastDeployment() (*Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	version, err := s.getVersion()
	if err != nil {
		return nil, err
	}

	key := s.makeKey("deploy", strconv.FormatUint(version, 10))
	return s.readDeployment(key)
}

// ReadDeployment retrieves a deployment from the store.
func (s *Persistence) ReadDeployment(version uint64) (*Deployment, error) {
	key := s.makeKey("deploy", strconv.FormatUint(version, 10))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.readDeployment(key)
}

func (s *Persistence) readDeployment(key string) (*Deployment, error) {
	kv, err := s.client.Get(s.ctx, key, readOptions)
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
func (s *Persistence) ListDeployments() ([]*Deployment, error) {
	key := s.makeKey("deploy", "")

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(s.ctx, key, readOptions)
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

func (s *Persistence) getVersion() (uint64, error) {
	key := s.makeKey("version", "last")
	kv, err := s.client.Get(s.ctx, key, readOptions)
	if err != nil {
		return 0, err
	}
	return s.unmarshalVersion(kv)
}

func (s *Persistence) putVersion(version uint64) error {
	key := s.makeKey("version", "last")
	b, _ := json.Marshal(version)
	return s.client.Put(s.ctx, key, b, writeOptions)
}

func (s *Persistence) putDeployment(d *Deployment) error {
	key := s.makeKey("deploy", strconv.FormatUint(d.Version(), 10))
	b, _ := json.Marshal(d)
	return s.client.Put(s.ctx, key, b, writeOptions)
}

func (s *Persistence) makeKey(prefix, id string) string {
	return fmt.Sprintf("%s$$$%s$$$%s%s", s.basePath, prefix, pathSeparator, id)
}

func (s *Persistence) mustMarshal(p *Persistence) []byte {
	b, _ := json.Marshal(p)
	return b
}

func (s *Persistence) unmarshalVersion(kv *store.KVPair) (uint64, error) {
	var v uint64
	if err := json.Unmarshal(kv.Value, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (s *Persistence) unmarshalDeployment(kv *store.KVPair) (*Deployment, error) {
	deploy := &Deployment{}
	if err := deploy.UnmarshalJSON(kv.Value); err != nil {
		return nil, err
	}
	return deploy, nil
}

func (s *Persistence) bugFix(in string) string {
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
