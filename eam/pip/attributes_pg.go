package pip

import (
	"context"
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

	sql := `INSERT INTO attribute
 (key,type,title,description,value,original,tags,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	now := db.now().UTC()
	params := []any{a.Key(), a.Type(), a.Title(), a.Description(), a.Value(), a.Original(), a.Tags(), now, user, now, user}

	if _, err := db.p.Exec(ctx, sql, params); err != nil {
		return nil, err
	}
	return a.WithAudit(now, user, now, user), nil
}

// ReadAttribute retrieves the identified attribute from the database.
func (db *PostgresDB) ReadAttribute(ctx context.Context, id string) (*models.Attribute, uint64, error) {
	sql := `SELECT key,type,title,description,value,original,tags,created,created_by,updated,updated_by
 FROM attribute
 WHERE key=$1`

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
func (db *PostgresDB) UpdateAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64, p *models.Attribute) (*models.Attribute, error) {
	user := convert.AnyToString(ctx.Value("user"))

	sql := `UPDATE attribute
 SET type=$3,title=$4,description=$5,value=$6,original=$7,tags=$8,updated=$9,updated_by=$10
 WHERE key=$1 AND updated=$2`

	now := db.now().UTC()
	params := []any{prev.Key(), timeFromLastIndex(lastIndex), p.Type(), p.Title(), p.Description(), p.Value(), p.Original(), p.Tags(), now, user}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("update failed; count=%d", count)
	}
	return attributeFromValues([]any{p.Key(), p.Type(), p.Title(), p.Description(), p.Value(), p.Original(), p.Tags(), p.Created(), p.CreatedBy(), now, user})
}

// DeleteAttribute removes an existing attribute from the database.
func (db *PostgresDB) DeleteAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64) (*models.Attribute, error) {
	sql := `DELETE attribute
 WHERE key=$1 AND updated=$2`
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
	sql := `SELECT key,type,title,description,value,original,tags,created,created_by,updated,updated_by FROM attribute`

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

func attributeFromValues(values []any) (*models.Attribute, error) {
	if len(values) != 11 {
		return nil, fmt.Errorf("invalid number of values")
	}

	a := models.NewOriginalAttribute(
		convert.AnyToString(values[0]), // key
		values[4],                      // value
		values[5],                      // original
		convert.AnyToString(values[1]), // type
	)

	return a.
		WithTitle(convert.AnyToString(values[2])).       // title
		WithDescription(convert.AnyToString(values[3])). // description
		WithTags(convert.AnyToStrings(values[6])...).    // tags
		WithAudit(
			convert.AnyToDateTime(values[7]).UTC(), // created
			convert.AnyToString(values[8]),         // createdBy
			convert.AnyToDateTime(values[9]).UTC(), // updated
			convert.AnyToString(values[10]),        // updatedBy
		), nil
}
