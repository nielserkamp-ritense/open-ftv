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

// NewTagDB instantiates a new PostgreSQL database connection for managing tags.
//
// The given context is used to signal a clean shutdown of the connection pool.
func NewTagDB(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*TagDB, error) {
	p, err := postgresql.New(ctx, dsn, maxLife, maxConn)
	if err != nil {
		return nil, err
	}
	return &TagDB{p: p, now: time.Now}, nil
}

// NewTagDBWithPool instantiates a new PostgreSQL database connection for managing tags using the given connection pool.
func NewTagDBWithPool(pool *postgresql.Postgres) *TagDB {
	return &TagDB{p: pool, now: time.Now}
}

// TagDB wraps a PostgreSQL connection pool with tag management functions.
type TagDB struct {
	p   *postgresql.Postgres
	now func() time.Time // for time-sensitive unit-tests.
}

// ListTags returns all policy languages from the database.
func (db *TagDB) ListTags(ctx context.Context) (policies.Tags, error) {
	sql := `SELECT tag,title,description,created,created_by,updated,updated_by FROM tag`

	list := make(policies.Tags, 0, 32)
	var err error

	err2 := db.p.Query(ctx, sql, nil, func(values []any) bool {
		var p *policies.Tag
		if p, err = tagFromValues(values); err == nil {
			list = append(list, *p)
		}
		return err == nil
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return list, nil
}

// ReplaceAllTags replaces all tags in the database with the given list.
func (db *TagDB) ReplaceAllTags(tags []*policies.Tag, user string) error {
	ctx := context.Background()

	_ = tags

	sql := "DELETE FROM tag"
	if _, err := db.p.Exec(ctx, sql, nil); err != nil {
		return err
	}

	now := time.Now().UTC()

	sql = "INSERT INTO tag (tag,title,description,created,created_by,updated,updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7)"
	for i := range tags {
		tag := tags[i]
		if _, err := db.p.Exec(ctx, sql, []any{tag.Id, tag.Name, tag.Description, now, user, now, user}); err != nil {
			return err
		}
	}
	return nil
}

func tagFromValues(values []any) (*policies.Tag, error) {
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

	return &policies.Tag{
		Id:          convert.AnyToString(values[0]),
		Name:        convert.AnyToString(values[1]),
		Description: convert.AnyToString(values[2]),
		Audit: policies.ObjectAudit{
			Created:   created,
			CreatedBy: convert.AnyToString(values[4]),
			Updated:   updated,
			UpdatedBy: convert.AnyToString(values[6]),
		},
	}, nil
}
