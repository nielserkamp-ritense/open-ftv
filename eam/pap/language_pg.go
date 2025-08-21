package pap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewLanguageDB instantiates a new PostgreSQL database connection for managing policy languages.
//
// The given context is used to signal a clean shutdown of the connection pool.
func NewLanguageDB(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*LanguageDB, error) {
	p, err := postgresql.New(ctx, dsn, maxLife, maxConn)
	if err != nil {
		return nil, err
	}
	return &LanguageDB{p: p, now: time.Now}, nil
}

// NewLanguageDBWithPool instantiates a new PostgreSQL database connection for managing policy languages using the given connection pool.
func NewLanguageDBWithPool(pool *postgresql.Postgres) *LanguageDB {
	return &LanguageDB{p: pool, now: time.Now}
}

// LanguageDB wraps a PostgreSQL connection pool with policy language management functions.
type LanguageDB struct {
	p   *postgresql.Postgres
	now func() time.Time // for time-sensitive unit-tests.
}

// ListLanguages returns all policy languages from the database.
func (db *LanguageDB) ListLanguages(ctx context.Context) (policies.Languages, error) {
	sql := `SELECT language,title,description,created,created_by,updated,updated_by FROM language`

	list := make(policies.Languages, 0, 32)
	var err error

	err2 := db.p.Query(ctx, sql, nil, func(values []any) bool {
		var p *policies.Language
		if p, err = languageFromValues(values); err == nil {
			list = append(list, *p)
		}
		return err == nil
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return list, nil
}

func languageFromValues(values []any) (*policies.Language, error) {
	if len(values) != 7 {
		return nil, fmt.Errorf("invalid number of values")
	}

	var created, updated string
	if d := convert.AnyToDateTime(values[3]); !d.IsZero() {
		created = d.Format(time.RFC3339Nano)
	}
	if d := convert.AnyToDateTime(values[5]); !d.IsZero() {
		updated = d.Format(time.RFC3339Nano)
	}

	return &policies.Language{
		Id:   convert.AnyToString(values[0]),
		Name: convert.AnyToString(values[1]),
		Audit: policies.ObjectAudit{
			Created:   created,
			CreatedBy: convert.AnyToString(values[4]),
			Updated:   updated,
			UpdatedBy: convert.AnyToString(values[6]),
		},
	}, nil
}
