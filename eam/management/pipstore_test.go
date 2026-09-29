package management

import (
	"fmt"
	"sync"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// memAttributes is an in-memory AttributeStore with the PIP's life cycle: the key-value PIP
// keeps neither status, tags nor attribution, and postgres is not a unit test.
type memAttributes struct {
	mu       sync.Mutex
	attrs    map[string]*models.Attribute
	versions map[string]map[int]*oas.AttributeVersion
	readErr  error
	lastBy   identity.Principal
}

func newMemAttributes() *memAttributes {
	return &memAttributes{attrs: map[string]*models.Attribute{}, versions: map[string]map[int]*oas.AttributeVersion{}}
}

func (m *memAttributes) IterateAttributes(f models.AttributeIterator) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, a := range m.attrs {
		f(a)
	}
}

func (m *memAttributes) GetAttribute(key string) (*models.Attribute, uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readErr != nil {
		return nil, 0, m.readErr
	}

	return m.attrs[key], 1, nil
}

func (m *memAttributes) GetAttributeAudit(string) ([]oas.AuditEntry, error)      { return nil, nil }
func (m *memAttributes) GetAttributeDeployments(string) ([]oas.UsageData, error) { return nil, nil }

func (m *memAttributes) GetAttributeVersions(key string) (oas.AttributeVersions, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(oas.AttributeVersions, 0, len(m.versions[key]))
	for _, v := range m.versions[key] {
		out = append(out, *v)
	}

	return out, nil
}

func (m *memAttributes) GetAttributeVersion(key string, version int) (*oas.AttributeVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.versions[key][version], nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memAttributes) AddAttributeFromOAS(in *oas.Attribute, user identity.Principal) (*models.Attribute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	a := models.NewAttributeFromOAS(in).WithAudit(time.Now(), user.ID, time.Now(), user.ID)
	m.attrs[in.Key], m.lastBy = a, user

	return a, nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memAttributes) UpdateAttributeStatus(key string, status models.Status, user identity.Principal) (*models.Attribute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.attrs[key]
	if prev == nil {
		return nil, fmt.Errorf("attribute not found")
	}

	if (prev.Status() == models.StatusConcept && status == models.StatusAccepted) ||
		(prev.Status() == models.StatusAccepted && status == models.StatusConcept) {
		m.lastBy = user
		return prev.WithStatus(status), nil
	}

	return nil, fmt.Errorf("%w: from %s to %s", pip.ErrInvalidStatusChange, prev.Status().String(), status.String())
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memAttributes) RestoreAttributeVersion(key string, version int, user identity.Principal) (*models.Attribute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.versions[key][version]
	if old == nil {
		return nil, fmt.Errorf("version %d of %s not found", version, key)
	}

	m.lastBy = user

	return m.attrs[key].RestoreFrom(old), nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memAttributes) RemoveAttribute(key string, user identity.Principal) (*models.Attribute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.attrs[key]
	delete(m.attrs, key)
	m.lastBy = user

	return prev, nil
}

// memEntities is memAttributes for entities.
type memEntities struct {
	mu       sync.Mutex
	entities map[string]*models.Entity
	versions map[string]map[int]*oas.EntityVersion
	readErr  error
	lastBy   identity.Principal
}

func newMemEntities() *memEntities {
	return &memEntities{entities: map[string]*models.Entity{}, versions: map[string]map[int]*oas.EntityVersion{}}
}

func (m *memEntities) IterateEntities(f models.EntityIterator) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, e := range m.entities {
		f(e)
	}
}

func (m *memEntities) GetEntity(uid string) (*models.Entity, uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readErr != nil {
		return nil, 0, m.readErr
	}

	return m.entities[uid], 1, nil
}

func (m *memEntities) GetEntityAudit(string) ([]oas.AuditEntry, error)      { return nil, nil }
func (m *memEntities) GetEntityDeployments(string) ([]oas.UsageData, error) { return nil, nil }

func (m *memEntities) GetEntityVersions(uid string) (oas.EntityVersions, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(oas.EntityVersions, 0, len(m.versions[uid]))
	for _, v := range m.versions[uid] {
		out = append(out, *v)
	}

	return out, nil
}

func (m *memEntities) GetEntityVersion(uid string, version int) (*oas.EntityVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.versions[uid][version], nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memEntities) AddEntityFromOAS(in *oas.Entity, user identity.Principal) (*models.Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	e := models.EntityFromOAS(in).WithAudit(time.Now(), user.ID, time.Now(), user.ID)
	m.entities[e.UID()], m.lastBy = e, user

	return e, nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memEntities) UpdateEntityStatus(uid string, status models.Status, user identity.Principal) (*models.Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.entities[uid]
	if prev == nil {
		return nil, fmt.Errorf("entity not found")
	}

	if (prev.Status() == models.StatusConcept && status == models.StatusAccepted) ||
		(prev.Status() == models.StatusAccepted && status == models.StatusConcept) {
		m.lastBy = user
		return prev.WithStatus(status), nil
	}

	return nil, fmt.Errorf("%w: from %s to %s", pip.ErrInvalidStatusChange, prev.Status().String(), status.String())
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memEntities) RestoreEntityVersion(uid string, version int, user identity.Principal) (*models.Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.versions[uid][version]
	if old == nil {
		return nil, fmt.Errorf("version %d of %s not found", version, uid)
	}

	m.lastBy = user

	return m.entities[uid].RestoreFrom(old), nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memEntities) RemoveEntity(uid string, user identity.Principal) (*models.Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.entities[uid]
	delete(m.entities, uid)
	m.lastBy = user

	return prev, nil
}
