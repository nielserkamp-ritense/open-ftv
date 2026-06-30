package pap

import (
	"context"
	"errors"
	"fmt"
	"time"

	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
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

// ListTags returns all tags from the database.
func (db *TagDB) ListTags(ctx context.Context) (oas.Tags, error) {
	sql := `SELECT tag,title,description,created,created_by,updated,updated_by FROM tag`

	list := make(oas.Tags, 0, 32)
	var err error

	err2 := db.p.Query(ctx, sql, nil, func(values []any) bool {
		var t *oas.Tag
		if t, err = tagFromValues(values); err == nil {
			list = append(list, *t)
		}
		return err == nil
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return list, nil
}

// CreateTag inserts the tag into the database.
func (db *TagDB) CreateTag(ctx context.Context, t *oas.Tag) (*oas.Tag, error) {
	now := time.Now().UTC()
	user := convert.AnyToString(ctx.Value("user"))

	sql := `INSERT INTO tag (tag,title,description,created,created_by,updated,updated_by) VALUES($1,$2,$3,$4,$5,$6,$7)`
	params := []any{t.Id, t.Name, t.Description, now, user, now, user}

	_, err := db.p.Exec(ctx, sql, params)
	if err != nil {
		return nil, err
	}

	t.Audit.Created = now.Format(time.RFC3339Nano)
	t.Audit.CreatedBy = user
	t.Audit.Updated = now.Format(time.RFC3339Nano)
	t.Audit.UpdatedBy = user
	return t, nil
}

// ReadTag returns the requested tag from the database.
func (db *TagDB) ReadTag(ctx context.Context, tag string) (*oas.Tag, uint64, error) {
	sql := `SELECT tag,title,description,created,created_by,updated,updated_by FROM tag WHERE tag=$1`

	var t *oas.Tag
	var err error

	err2 := db.p.Query(ctx, sql, []any{tag}, func(values []any) bool {
		t, err = tagFromValues(values)
		return false // only one possible.
	})

	if err != nil || err2 != nil {
		return nil, 0, errors.Join(err, err2)
	}

	updated, _ := time.Parse(time.RFC3339Nano, t.Audit.Updated)
	return t, timeToLastIndex(updated), nil
}

// UpdateTag replaces the tag in the database.
func (db *TagDB) UpdateTag(ctx context.Context, prev *oas.Tag, lastIndex uint64, t *oas.Tag) (*oas.Tag, error) {
	now := time.Now().UTC()
	user := convert.AnyToString(ctx.Value("user"))

	sql := `UPDATE tag SET title=$3,description=$4,updated=$5,updated_by=$6 WHERE tag=$1 AND updated=$2`
	params := []any{prev.Id, timeFromLastIndex(lastIndex), t.Name, t.Description, now, user}

	_, err := db.p.Exec(ctx, sql, params)
	if err != nil {
		return nil, err
	}

	t.Audit.Updated = now.Format(time.RFC3339Nano)
	t.Audit.UpdatedBy = user
	return t, nil
}

// DeleteTag removes the tag from the database.
func (db *TagDB) DeleteTag(ctx context.Context, prev *oas.Tag, lastIndex uint64) (*oas.Tag, error) {
	sql := `DELETE tag WHERE tag=$1 AND updated = $2`
	params := []any{prev.Id, timeFromLastIndex(lastIndex)}

	_, err := db.p.Exec(ctx, sql, params)
	if err != nil {
		return nil, err
	}
	return prev, nil
}

// EnsureTags inserts tags that are not already present.
func (db *TagDB) EnsureTags(tags []*oas.Tag, user string) error {
	ctx := context.Background()
	now := db.now().UTC()
	sql := `INSERT INTO tag (tag,title,description,created,created_by,updated,updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (tag) DO NOTHING`

	for i := range tags {
		tag := tags[i]
		if tag == nil || tag.Id == "" {
			continue
		}
		if _, err := db.p.Exec(ctx, sql, []any{tag.Id, tag.Name, tag.Description, now, user, now, user}); err != nil {
			return err
		}
	}
	return nil
}

func tagFromValues(values []any) (*oas.Tag, error) {
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

	return &oas.Tag{
		Id:          convert.AnyToString(values[0]),
		Name:        convert.AnyToString(values[1]),
		Description: convert.AnyToString(values[2]),
		Audit: oas.ObjectAudit{
			Created:   created,
			CreatedBy: convert.AnyToString(values[4]),
			Updated:   updated,
			UpdatedBy: convert.AnyToString(values[6]),
		},
	}, nil
}
