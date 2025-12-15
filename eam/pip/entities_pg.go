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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// CreateEntity creates a new entity in the database.
func (db *PostgresDB) CreateEntity(ctx context.Context, e *models.Entity) (*models.Entity, error) {
	user := convert.AnyToString(ctx.Value("user"))

	sql := `INSERT INTO entity
 (status,type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	now := db.now().UTC()
	params := []any{e.StatusName(), e.Type(), e.ID(), e.Title(), e.Description(), e.Tags(), gobEncodeAttributes(e.Attributes()), e.Parents(), now, user, now, user}

	if _, err := db.p.Exec(ctx, sql, params); err != nil {
		return nil, err
	}
	return e.WithAudit(now, user, now, user), nil
}

// ReadEntity retrieves the identified entity from the database.
func (db *PostgresDB) ReadEntity(ctx context.Context, ns, id string) (*models.Entity, uint64, error) {
	sql := `SELECT status,type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by
 FROM entity
 WHERE type=$1 AND id=$2`

	params := []any{ns, id}

	var e *models.Entity
	var err error

	err2 := db.p.Query(ctx, sql, params, func(values []any) bool {
		e, err = entityFromValues(values)
		return false // there can only be one!
	})

	if err != nil || err2 != nil {
		return nil, 0, errors.Join(err, err2)
	}

	if e == nil {
		return nil, 0, nil
	}
	return e, timeToLastIndex(e.Updated()), nil
}

// ReadEntityAudit retrieves the audit-log for the identified policy from the database.
func (db *PostgresDB) ReadEntityAudit(ctx context.Context, ns, id string) ([]oas.AuditEntry, error) {
	sql := `SELECT created,operation,user_id FROM entity_audit WHERE type=$1 AND id=$2 ORDER BY created DESC LIMIT 100`
	params := []any{ns, id}
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

// ReadEntityDeployments retrieves the deployment-log for the identified policy from the database.
func (db *PostgresDB) ReadEntityDeployments(ctx context.Context, ns, id string) ([]oas.UsageData, error) {
	sql := `SELECT b.bundle_id, d.version, d.created, d.created_by, d.status
 FROM entity_deployment e
 INNER JOIN deployment_bundle b ON b.id=e.bundle_id
 INNER JOIN deployment d ON d.version=b.version
 WHERE e.type=$1 AND a.id=$2
 ORDER BY b.created DESC
 LIMIT 100`

	params := []any{ns, id}
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

// ReadEntityVersions retrieves the versions for the identified attribute from the database.
func (db *PostgresDB) ReadEntityVersions(ctx context.Context, ns, id string) (oas.EntityVersions, error) {
	sql := `SELECT version,type,id,title,description,tags,attributes,parents
 FROM entity_version WHERE type=$1 AND id=$2 ORDER BY version DESC`

	params := []any{ns, id}
	out := make([]oas.EntityVersion, 0)

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		out = append(out, entityVersionFromDB(values))
		return true
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadEntityVersion retrieves a specific version for the identified attribute from the database.
func (db *PostgresDB) ReadEntityVersion(ctx context.Context, ns, id string, version int) (*oas.EntityVersion, error) {
	sql := `SELECT version,type,id,title,description,tags,attributes,parents
 FROM attribute_version WHERE version=$1 AND type=$2 AND id=$3 ORDER BY version DESC`

	params := []any{version, ns, id}
	var out *oas.EntityVersion

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		pol := entityVersionFromDB(values)
		out = &pol
		return false // there can only be one
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

func entityVersionFromDB(values []any) oas.EntityVersion {
	attrs, _ := gobDecodeAttributes(values[6].([]byte))

	return oas.EntityVersion{
		Version:    int(convert.AnyToInt64(values[0])),
		Type:       postgresql.AnyToUUID(values[1]),
		Id:         convert.AnyToString(values[2]),
		Attributes: attrs.ToOAS(),
		Metadata: oas.Metadata{
			Title:       convert.AnyToString(values[3]),
			Description: convert.AnyToString(values[4]),
			Tags:        convert.AnyToStrings(values[5]),
		},
	}
}

// UpdateEntity replaces an existing entity in the database.
func (db *PostgresDB) UpdateEntity(ctx context.Context, prev *models.Entity, lastIndex uint64, e *models.Entity) (*models.Entity, error) {
	user := convert.AnyToString(ctx.Value("user"))

	sql := `UPDATE entity
 SET title=$4,description=$5,tags=$6,attributes=$7,parents=$8,status=$9,updated=$10,updated_by=$11
 WHERE type=$1 AND id=$2 AND updated=$3`

	attr := gobEncodeAttributes(e.Attributes())
	now := db.now().UTC()
	params := []any{prev.Type(), prev.ID(), timeFromLastIndex(lastIndex), e.Title(), e.Description(), e.Tags(), attr, e.Parents(), e.StatusName(), now, user}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("update failed; count=%d", count)
	}

	return e.WithAudit(e.Created(), e.CreatedBy(), now, user), nil
}

// DeleteEntity removes an existing entity from the database.
func (db *PostgresDB) DeleteEntity(ctx context.Context, prev *models.Entity, lastIndex uint64) (*models.Entity, error) {
	sql := `DELETE entity
 WHERE type=$1 AND id=$2 AND updated=$3`
	params := []any{prev.Type(), prev.ID(), timeFromLastIndex(lastIndex)}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("delete failed; count=%d", count)
	}
	return prev, nil
}

// ListEntities returns entities from the database.
func (db *PostgresDB) ListEntities(ctx context.Context) ([]*models.Entity, error) {
	sql := `SELECT status,type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by FROM entity`

	list := make([]*models.Entity, 0, 32)
	var err error

	err2 := db.p.Query(ctx, sql, nil, func(values []any) bool {
		var p *models.Entity
		if p, err = entityFromValues(values); err == nil {
			list = append(list, p)
		}
		return err == nil
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return list, nil
}

func entityFromValues(values []any) (*models.Entity, error) {
	if len(values) != 12 {
		return nil, fmt.Errorf("invalid number of values")
	}

	attrs, err := gobDecodeAttributes(values[6].([]byte))
	if err != nil {
		return nil, err
	}

	e := models.NewEntity(
		convert.AnyToString(values[1]),     // type
		convert.AnyToString(values[2]),     // id
		attrs,                              // attributes
		convert.AnyToStrings(values[7])..., // parents
	)

	return e.
		WithStatus(models.StatusFromString(convert.AnyToString(values[0]))). // status
		WithTitle(convert.AnyToString(values[3])).                           // title
		WithDescription(convert.AnyToString(values[4])).                     // description
		WithTags(convert.AnyToStrings(values[5])...).                        // tags
		WithAudit(
			convert.AnyToDateTime(values[8]).UTC(),  // created
			convert.AnyToString(values[9]),          // createdBy
			convert.AnyToDateTime(values[10]).UTC(), // updated
			convert.AnyToString(values[11]),         // updatedBy
		), nil
}
