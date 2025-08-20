package pip

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// CreateEntity creates a new entity in the database.
func (db *PostgresDB) CreateEntity(ctx context.Context, user string, e *models.Entity) (*models.Entity, error) {
	sql := `INSERT INTO entity
 (type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	now := db.now().UTC()
	params := []any{e.Type(), e.ID(), e.Title(), e.Description(), e.Tags(), gobEncodeAttributes(e.Attributes()), e.Parents(), now, user, now, user}

	if _, err := db.p.Exec(ctx, sql, params); err != nil {
		return nil, err
	}
	return e.WithAudit(now, user, now, user), nil
}

// ReadEntity retrieves the identified entity from the database.
func (db *PostgresDB) ReadEntity(ctx context.Context, ns, id string) (*models.Entity, uint64, error) {
	sql := `SELECT type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by
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

// UpdateEntity replaces an existing entity in the database.
func (db *PostgresDB) UpdateEntity(ctx context.Context, user string, prev *models.Entity, lastIndex uint64, p *models.Entity) (*models.Entity, error) {
	sql := `UPDATE entity
 SET title=$4,description=$5,tags=$6,attributes=$7,parents=$8,updated=$9,updated_by=$10
 WHERE type=$1 AND id=$2 AND updated=$3`

	attr := gobEncodeAttributes(p.Attributes())
	now := db.now().UTC()
	params := []any{prev.Type(), prev.ID(), timeFromLastIndex(lastIndex), p.Title(), p.Description(), p.Tags(), attr, p.Parents(), now, user}

	if count, err := db.p.Exec(ctx, sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("update failed; count=%d", count)
	}
	return entityFromValues([]any{p.Type(), p.ID(), p.Title(), p.Description(), p.Tags(), attr, p.Parents(), p.Created(), p.CreatedBy(), now, user})
}

// DeleteEntity removes an existing entity from the database.
func (db *PostgresDB) DeleteEntity(ctx context.Context, _ string, prev *models.Entity, lastIndex uint64) (*models.Entity, error) {
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
	sql := `SELECT type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by FROM entity`

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
	if len(values) != 11 {
		return nil, fmt.Errorf("invalid number of values")
	}

	attrs, err := gobDecodeAttributes(values[5].([]byte))
	if err != nil {
		return nil, err
	}

	a := models.NewEntity(
		convert.AnyToString(values[0]),     // type
		convert.AnyToString(values[1]),     // id
		attrs,                              // attributes
		convert.AnyToStrings(values[6])..., // parents
	)

	return a.
		WithTitle(convert.AnyToString(values[2])).       // title
		WithDescription(convert.AnyToString(values[3])). // description
		WithTags(convert.AnyToStrings(values[4])...).    // tags
		WithAudit(
			convert.AnyToDateTime(values[7]).UTC(), // created
			convert.AnyToString(values[8]),         // createdBy
			convert.AnyToDateTime(values[9]).UTC(), // updated
			convert.AnyToString(values[10]),        // updatedBy
		), nil
}
