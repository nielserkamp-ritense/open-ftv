package pip

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// CreateAttribute creates a new attribute in the database.
func (db *PostgresDB) CreateAttribute(ctx context.Context, a *models.Attribute) (*models.Attribute, error) {
	user := convert.AnyToString(ctx.Value("user"))
	v, o, err := encodeValues(a)
	if err != nil {
		return nil, err
	}

	sql := `INSERT INTO attribute
 (status,key,type,title,description,value,original,tags,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	now := db.now().UTC()
	params := []any{a.StatusName(), a.Key(), a.Type(), a.Title(), a.Description(), v, o, a.Tags(), now, user, now, user}

	if _, err = db.p.Exec(ctx, sql, params); err != nil {
		return nil, err
	}
	return a.WithAudit(now, user, now, user), nil
}

// ReadAttribute retrieves the identified attribute from the database.
func (db *PostgresDB) ReadAttribute(ctx context.Context, id string) (*models.Attribute, uint64, error) {
	sql := `SELECT status,key,type,title,description,value,original,tags,created,created_by,updated,updated_by
 FROM attribute WHERE key=$1`

	params := []any{id}

	var a *models.Attribute
	var err error

	err2 := db.p.Query(ctx, sql, params, func(values []any) bool {
		a, err = attributeFromValues(values)
		return false // there can only be one!
	})

	if err != nil || err2 != nil {
		return nil, 0, errors.Join(err, err2)
	}

	if a == nil {
		return nil, 0, nil
	}
	return a, timeToLastIndex(a.Updated()), nil
}

// ReadAttributeAudit retrieves the audit-log for the identified policy from the database.
func (db *PostgresDB) ReadAttributeAudit(ctx context.Context, key string) ([]oas.AuditEntry, error) {
	sql := `SELECT created,operation,user_id FROM attribute_audit WHERE key=$1 ORDER BY created DESC LIMIT 100`
	params := []any{key}
	out := make([]oas.AuditEntry, 0)

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		out = append(out, oas.AuditEntry{
			Created:   convert.AnyToDateTime(values[0]).Format(time.RFC3339),
			Operation: convert.AnyToString(values[1]),
			UserId:    convert.AnyToString(values[2]),
		})
		return true
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadAttributeDeployments retrieves the deployment-log for the identified policy from the database.
func (db *PostgresDB) ReadAttributeDeployments(ctx context.Context, key string) ([]oas.UsageData, error) {
	sql := `SELECT b.bundle_id, d.version, d.created, d.created_by, d.status
 FROM attribute_deployment a
 INNER JOIN deployment_bundle b ON b.id=a.bundle_id
 INNER JOIN deployment d ON d.version=b.version
 WHERE a.key=$1
 ORDER BY b.created DESC
 LIMIT 100`

	params := []any{key}
	out := make([]oas.UsageData, 0)

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		out = append(out, oas.UsageData{
			Bundle:    convert.AnyToString(values[0]),
			Version:   int(convert.AnyToInt64(values[1])),
			Created:   convert.AnyToString(values[2]),
			CreatedBy: convert.AnyToString(values[3]),
			Status:    bundles.Status(convert.AnyToInt64(values[4])).Status(),
		})
		return true
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateAttribute replaces an existing attribute in the database.
func (db *PostgresDB) UpdateAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64, a *models.Attribute) (*models.Attribute, error) {
	user := convert.AnyToString(ctx.Value("user"))

	v, o, err := encodeValues(a)
	if err != nil {
		return nil, err
	}

	sql := `UPDATE attribute
 SET type=$3,title=$4,description=$5,value=$6,original=$7,tags=$8,status=$9,updated=$10,updated_by=$11
 WHERE key=$1 AND updated=$2`

	now := db.now().UTC()
	params := []any{prev.Key(), timeFromLastIndex(lastIndex), a.Type(), a.Title(), a.Description(), v, o, a.Tags(), a.StatusName(), now, user}

	if count, err2 := db.p.Exec(ctx, sql, params); err2 != nil || count != 1 {
		if err2 != nil {
			return nil, err2
		}
		return nil, fmt.Errorf("update failed; count=%d", count)
	}

	return a.WithAudit(a.Created(), a.CreatedBy(), now, user), nil
}

// DeleteAttribute removes an existing attribute from the database.
func (db *PostgresDB) DeleteAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64) (*models.Attribute, error) {
	sql := `DELETE attribute WHERE key=$1 AND updated=$2`
	params := []any{prev.Key(), timeFromLastIndex(lastIndex)}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("delete failed; count=%d", count)
	}
	return prev, nil
}

// ListAttributes returns attributes from the database.
func (db *PostgresDB) ListAttributes(ctx context.Context) ([]*models.Attribute, error) {
	sql := `SELECT status,key,type,title,description,value,original,tags,created,created_by,updated,updated_by FROM attribute`

	list := make([]*models.Attribute, 0, 32)
	var err error

	err2 := db.p.Query(ctx, sql, nil, func(values []any) bool {
		var p *models.Attribute
		if p, err = attributeFromValues(values); err == nil {
			list = append(list, p)
		}
		return err == nil
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return list, nil
}

func encodeValues(a *models.Attribute) ([]byte, []byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	q := a.Value()
	if err := enc.Encode(&q); err != nil {
		return nil, nil, fmt.Errorf("unable to encode value [%T]: %w", q, err)
	}
	v := buf.Bytes()

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	q = a.Original()
	if err := enc.Encode(&q); err != nil {
		return nil, nil, fmt.Errorf("unable to encode original value [%T]: %w", q, err)
	}

	return v, buf.Bytes(), nil
}

func attributeFromValues(values []any) (*models.Attribute, error) {
	if len(values) != 12 {
		return nil, fmt.Errorf("invalid number of values")
	}

	v, o := decodeValues(values[5], values[6])

	a := models.NewOriginalAttribute(
		convert.AnyToString(values[1]), // key
		v,                              // value
		o,                              // original
		convert.AnyToString(values[2]), // type
	)

	return a.
		WithStatus(models.StatusFromString(convert.AnyToString(values[0]))). // status
		WithTitle(convert.AnyToString(values[3])).                           // title
		WithDescription(convert.AnyToString(values[4])).                     // description
		WithTags(convert.AnyToStrings(values[7])...).                        // tags
		WithAudit(
			convert.AnyToDateTime(values[8]).UTC(),  // created
			convert.AnyToString(values[9]),          // createdBy
			convert.AnyToDateTime(values[10]).UTC(), // updated
			convert.AnyToString(values[11]),         // updatedBy
		), nil
}

func decodeValues(in1, in2 any) (any, any) {
	var b1, b2 []byte
	var ok1, ok2 bool

	var v, o any

	if b1, ok1 = in1.([]byte); !ok1 {
		v = in1
	}
	if b2, ok2 = in2.([]byte); !ok2 {
		o = in2
	}

	buf := &bytes.Buffer{}
	dec := gob.NewDecoder(buf)

	if ok1 {
		buf.Write(b1)
		if err := dec.Decode(&v); err != nil {
			v = nil
		}
	}

	if ok2 {
		buf.Reset()
		buf.Write(b2)
		if err := dec.Decode(&o); err != nil {
			o = nil
		}
	}

	return v, o
}
