package bundles

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewPostgresDB instantiates a new bundle deployment persistence manager with a PostgreSQL database as backend.
func NewPostgresDB(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*PostgresDB, error) {
	db, err := postgresql.New(ctx, dsn, maxLife, maxConn)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{db: db}, nil
}

// NewPostgresDBWithPool instantiates a new bundle deployment persistence manager with the given PostgreSQL connection pool.
func NewPostgresDBWithPool(db *postgresql.Postgres) *PostgresDB {
	return &PostgresDB{db: db}
}

// PostgresDB contains the details to manage bundle deployments using a KV store as backend.
type PostgresDB struct {
	db    *postgresql.Postgres
	mutex sync.RWMutex
}

// Generate creates a new deployment in the store.
//
// If the last deployment is still active, this function will return an error.
func (s *PostgresDB) Generate(ctx context.Context, description, user string) (*Deployment, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion(ctx)
	if err != nil && err.Error() != "key not found in store" {
		return nil, err
	}

	if err == nil {
		deploy, err2 := s.readDeployment(ctx, version)
		switch {
		case err2 != nil:
			return nil, err2
		case deploy != nil:
			if status := deploy.Status(); status < Failed {
				return nil, fmt.Errorf("deployment %d is still active [%s]", version, status.String())
			}
		}
	}

	version++
	deploy := NewDeployment(version, description, user)

	if err = s.createDeployment(ctx, deploy); err != nil {
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
func (s *PostgresDB) Advance() (*Deployment, error) {
	ctx := context.Background()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	deploy, err3 := s.readDeployment(ctx, version)
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

	if err = s.updateDeployment(ctx, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// Fail sets the status of the last deployment in the store to *Failed*.
//
// If the last deployment is no longer active, this function will return an error.
func (s *PostgresDB) Fail(msg string) (*Deployment, error) {
	ctx := context.Background()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	version, err := s.getVersion(ctx)
	if err != nil {
		return nil, err
	}

	deploy, err3 := s.readDeployment(ctx, version)
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

	if err = s.updateDeployment(ctx, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

// LastDeployment retrieves the last deployment from the store.
func (s *PostgresDB) LastDeployment(ctx context.Context) (*Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	version, err := s.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	return s.readDeployment(ctx, version)
}

// ReadDeployment retrieves a deployment from the store.
func (s *PostgresDB) ReadDeployment(ctx context.Context, version uint64) (*Deployment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.readDeployment(ctx, version)
}

func (s *PostgresDB) readDeployment(ctx context.Context, version uint64) (*Deployment, error) {
	sql := "SELECT version,status,title,description,message,created,created_by,updated,updated_by FROM deployment WHERE version=$1"

	var out *Deployment
	if err := s.db.Query(ctx, sql, []any{version}, func(values []any) bool {
		out = deploymentFromValues(values)
		return false // only one record.
	}); err != nil {
		return nil, err
	}

	return out, nil
}

// ListDeployments retrieves all deployments from the store.
func (s *PostgresDB) ListDeployments(ctx context.Context) ([]*Deployment, error) {
	sql := "SELECT version,status,title,description,message,created,created_by,updated,updated_by FROM deployment"

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	out := make([]*Deployment, 0)
	if err := s.db.Query(ctx, sql, nil, func(values []any) bool {
		out = append(out, deploymentFromValues(values))
		return true
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *PostgresDB) getVersion(ctx context.Context) (uint64, error) {
	sql := "SELECT MAX(version) FROM deployment"

	var version uint64
	if err := s.db.Query(ctx, sql, nil, func(values []any) bool {
		if len(values) == 1 {
			version = convert.AnyToUint64(values[0])
		}
		return false // only one record.
	}); err != nil {
		return 0, err
	}

	return version, nil
}

func (s *PostgresDB) createDeployment(ctx context.Context, d *Deployment) error {
	sql := "INSERT INTO deployment (version,status,title,description,created,created_by,updated,updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)"
	params := []any{int(d.version), int(d.status), d.title, d.description, d.created, d.createdBy, d.updated, d.updatedBy}

	_, err := s.db.Exec(ctx, sql, params)
	return err
}

func (s *PostgresDB) updateDeployment(ctx context.Context, d *Deployment) error {
	sql := "UPDATE deployment SET status=$2,message=$3,updated=$4,updated_by=$5 WHERE version=$1"
	params := []any{int(d.version), int(d.status), d.msg, d.updated, d.updatedBy}

	_, err := s.db.Exec(ctx, sql, params)
	return err
}

func deploymentFromValues(values []any) *Deployment {
	return &Deployment{
		version:     convert.AnyToUint64(values[0]),
		status:      Status(convert.AnyToInt64(values[1])),
		title:       convert.AnyToString(values[2]),
		description: convert.AnyToString(values[3]),
		msg:         convert.AnyToString(values[4]),
		created:     convert.AnyToDateTime(values[5]),
		createdBy:   convert.AnyToString(values[6]),
		updated:     convert.AnyToDateTime(values[7]),
		updatedBy:   convert.AnyToString(values[8]),
	}
}
