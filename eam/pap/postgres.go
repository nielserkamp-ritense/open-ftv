package pap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewPostgresDB instantiates a new PostgreSQL database connection handler.
//
// The given context is used to signal a clean shutdown of the connection pool.
func NewPostgresDB(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*PostgresDB, error) {
	p, err := postgresql.New(ctx, dsn, maxLife, maxConn)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{p: p, now: time.Now}, nil
}

// PostgresDB wraps a PostgreSQL connection pool with policy management functions.
type PostgresDB struct {
	p   *postgresql.Postgres
	now func() time.Time
}

// CreatePolicy creates a new policy into the database.
func (db *PostgresDB) CreatePolicy(ctx context.Context, user string, p *models.Policy) (*models.Policy, error) {
	sql := `INSERT INTO policy
 (language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	now := db.now().UTC()
	params := []any{p.Language(), p.ID(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), now, user, now, user}

	if _, err := db.p.Exec(ctx, sql, params); err != nil {
		return nil, err
	}
	return p.WithAudit(now, user, now, user), nil
}

// ReadPolicy retrieves the identified policy from the database.
func (db *PostgresDB) ReadPolicy(ctx context.Context, id string) (*models.Policy, uint64, error) {
	sql := `SELECT language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by
 FROM policy
 WHERE id=$1`

	params := []any{id}

	var p *models.Policy
	var err error

	err2 := db.p.Query(ctx, sql, params, func(values []any) bool {
		p, err = policyFromValues(values)
		return false // there can only be one!
	})

	if err != nil || err2 != nil {
		return nil, 0, errors.Join(err, err2)
	}
	return p, timeToLastIndex(p.Updated()), nil
}

// UpdatePolicy replaces an existing policy in the database.
func (db *PostgresDB) UpdatePolicy(ctx context.Context, user string, prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error) {
	sql := `UPDATE policy
 SET language=$3,title=$4,description=$5,rvva_id=$6,uri=$7,tags=$8,content=$9,updated=$10,updated_by=$11
 WHERE id=$1 AND updated=$2`

	now := db.now().UTC()
	params := []any{prev.ID(), timeFromLastIndex(lastIndex), p.Language(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), now, user}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("update failed; count=%d", count)
	}
	return policyFromValues([]any{p.Language(), p.ID(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), p.Created(), p.CreatedBy(), now, user})
}

// DeletePolicy removes an existing policy from the database.
func (db *PostgresDB) DeletePolicy(ctx context.Context, _ string, prev *models.Policy, lastIndex uint64) (*models.Policy, error) {
	sql := `DELETE policy
 WHERE id=$1 AND updated=$2`
	params := []any{prev.ID(), timeFromLastIndex(lastIndex)}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("delete failed; count=%d", count)
	}
	return prev, nil
}

// ListPolicies returns policies from the database, optionally limited to the given policy language.
//
// If *language* is empty, all policies in the database will be returned.
func (db *PostgresDB) ListPolicies(ctx context.Context, language string) ([]*models.Policy, error) {
	sql := `SELECT language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by FROM policy`

	var params []any
	if language != "" {
		sql = fmt.Sprintf("%s WHERE language=$1", sql)
		params = append(params, language)
	}

	list := make([]*models.Policy, 0, 32)
	var err error

	err2 := db.p.Query(ctx, sql, params, func(values []any) bool {
		var p *models.Policy
		if p, err = policyFromValues(values); err == nil {
			list = append(list, p)
		}
		return err == nil
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return list, nil
}

func policyFromValues(values []any) (*models.Policy, error) {
	if len(values) != 12 {
		return nil, fmt.Errorf("invalid number of values")
	}

	p, err := models.NewPolicyFromData(
		postgresql.AnyToUUID(values[1]),                       // id
		convert.AnyToString(values[0]),                        // language
		convert.AnyToString(values[4]),                        // rvvaID
		convert.AnyToString(values[5]),                        // uri
		bytes.NewBufferString(convert.AnyToString(values[7])), // content
	)
	if err != nil {
		return nil, err
	}

	return p.
		WithTitle(convert.AnyToString(values[2])).       // title
		WithDescription(convert.AnyToString(values[3])). // description
		WithTags(convert.AnyToStrings(values[6])...).    // tags
		WithAudit(
			convert.AnyToDateTime(values[8]).UTC(),  // created
			convert.AnyToString(values[9]),          // createdBy
			convert.AnyToDateTime(values[10]).UTC(), // updated
			convert.AnyToString(values[11]),         // updatedBy
		), nil
}

func timeToLastIndex(t time.Time) uint64 {
	if t.IsZero() {
		return 0
	}
	return uint64(t.UnixNano())
}

func timeFromLastIndex(lastIndex uint64) time.Time {
	if lastIndex == 0 {
		return time.Time{}
	}
	i := int64(lastIndex)
	j := int64(time.Second)
	return time.Unix(i/j, i%j).UTC()
}
