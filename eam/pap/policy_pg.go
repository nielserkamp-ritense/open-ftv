package pap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewPolicyDB instantiates a new PostgreSQL database connection for managing policies.
//
// The given context is used to signal a clean shutdown of the connection pool.
func NewPolicyDB(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*PolicyDB, error) {
	p, err := postgresql.New(ctx, dsn, maxLife, maxConn)
	if err != nil {
		return nil, err
	}
	return &PolicyDB{p: p, now: time.Now}, nil
}

// NewPolicyDBWithPool instantiates a new PostgreSQL database connection for managing policies using the given connection pool.
func NewPolicyDBWithPool(pool *postgresql.Postgres) *PolicyDB {
	return &PolicyDB{p: pool, now: time.Now}
}

// PolicyDB wraps a PostgreSQL connection pool with policy management functions.
type PolicyDB struct {
	p   *postgresql.Postgres
	now func() time.Time // for time-sensitive unit-tests.
}

// CreatePolicy creates a new policy into the database.
func (db *PolicyDB) CreatePolicy(ctx context.Context, user identity.Principal, p *models.Policy) (*models.Policy, error) {
	now := db.now().UTC()

	sql := `INSERT INTO policy
 (status,language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	by := user.ID
	params := []any{p.StatusName(), p.Language(), p.ID(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), now, by, now, by}

	if _, err := db.p.Exec(identity.WithContext(ctx, user), sql, params); err != nil {
		return nil, err
	}

	return p.WithAudit(now, by, now, by), nil
}

// ReadPolicy retrieves the identified policy from the database.
func (db *PolicyDB) ReadPolicy(ctx context.Context, id string) (*models.Policy, uint64, error) {
	sql := `SELECT status,language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by
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

	if p == nil {
		return nil, 0, nil
	}
	return p, timeToLastIndex(p.Updated()), nil
}

// ReadPolicyAudit retrieves the audit-log for the identified policy from the database.
func (db *PolicyDB) ReadPolicyAudit(ctx context.Context, id string) ([]oas.AuditEntry, error) {
	sql := `SELECT created,operation,user_id FROM policy_audit WHERE id=$1 ORDER BY created DESC LIMIT 100`
	params := []any{id}
	out := make([]oas.AuditEntry, 0)

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		out = append(out, oas.AuditEntry{
			Created:   convert.AnyToDateTime(values[0]).Format(time.RFC3339),
			Operation: convert.AnyToString(values[1]),
			User:      oas.Principal{Id: convert.AnyToString(values[2])},
		})
		return true
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadPolicyDeployments retrieves the deployment-log for the identified policy from the database.
func (db *PolicyDB) ReadPolicyDeployments(ctx context.Context, id string) ([]oas.UsageData, error) {
	sql := `SELECT b.bundle_id, d.version, d.created, d.created_by, d.status
 FROM policy_deployment p
 INNER JOIN deployment_bundle b ON b.id=p.bundle_id
 INNER JOIN deployment d ON d.version=b.version
 WHERE p.id=$1
 ORDER BY b.created DESC
 LIMIT 100`

	params := []any{id}
	out := make([]oas.UsageData, 0)

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		out = append(out, oas.UsageData{
			Bundle:    convert.AnyToString(values[0]),
			Version:   int(convert.AnyToInt64(values[1])),
			Created:   convert.AnyToString(values[2]),
			CreatedBy: optionalPrincipal(convert.AnyToString(values[3])),
			Status:    bundles.Status(convert.AnyToInt64(values[4])).Status(),
		})
		return true
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadPolicyVersions retrieves the versions for the identified policy from the database.
func (db *PolicyDB) ReadPolicyVersions(ctx context.Context, id string) (oas.PolicyVersions, error) {
	sql := `SELECT version,language,id,title,description,rvva_id,uri,tags,content
 FROM policy_version WHERE id=$1 ORDER BY version DESC`

	params := []any{id}
	out := make([]oas.PolicyVersion, 0)

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		out = append(out, polVersionFromDB(values))
		return true
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadPolicyVersion retrieves a specific version for the identified policy from the database.
func (db *PolicyDB) ReadPolicyVersion(ctx context.Context, id string, version int) (*oas.PolicyVersion, error) {
	sql := `SELECT version,language,id,title,description,rvva_id,uri,tags,content
 FROM policy_version WHERE version=$1 AND id=$2 ORDER BY version DESC`

	params := []any{version, id}
	var out *oas.PolicyVersion

	err := db.p.Query(ctx, sql, params, func(values []any) bool {
		pol := polVersionFromDB(values)
		out = &pol
		return false
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

func polVersionFromDB(values []any) oas.PolicyVersion {
	return oas.PolicyVersion{
		Version:  int(convert.AnyToInt64(values[0])),
		Language: convert.AnyToString(values[1]),
		Id:       postgresql.AnyToUUID(values[2]),
		Metadata: oas.Metadata{
			Title:       convert.AnyToString(values[3]),
			Description: convert.AnyToString(values[4]),
			RvvaId:      convert.AnyToString(values[5]),
			Url:         convert.AnyToString(values[6]),
			Tags:        convert.AnyToStrings(values[7]),
		},
		Data: convert.AnyToString(values[8]),
	}
}

// UpdatePolicy replaces an existing policy in the database.
func (db *PolicyDB) UpdatePolicy(ctx context.Context, user identity.Principal, prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error) {
	now := db.now().UTC()

	sql := `UPDATE policy
 SET language=$3,title=$4,description=$5,rvva_id=$6,uri=$7,tags=$8,content=$9,status=$10,updated=$11,updated_by=$12
 WHERE id=$1 AND updated=$2`

	by := user.ID
	params := []any{prev.ID(), timeFromLastIndex(lastIndex), p.Language(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), p.StatusName(), now, by}

	if count, err := db.p.Exec(identity.WithContext(ctx, user), sql, params); err != nil || count != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("update failed; count=%d", count)
	}

	return p.WithAudit(p.Created(), p.CreatedBy(), now, by), nil
}

// DeletePolicy removes an existing policy from the database.
func (db *PolicyDB) DeletePolicy(ctx context.Context, user identity.Principal, prev *models.Policy, lastIndex uint64) (*models.Policy, error) {
	sql := `DELETE FROM policy WHERE id=$1 AND updated=$2`
	params := []any{prev.ID(), timeFromLastIndex(lastIndex)}

	if count, err := db.p.Exec(identity.WithContext(ctx, user), sql, params); err != nil || count != 1 {
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
func (db *PolicyDB) ListPolicies(ctx context.Context, language string) ([]*models.Policy, error) {
	sql := `SELECT status,language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by FROM policy`

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
	if len(values) != 13 {
		return nil, fmt.Errorf("invalid number of values")
	}

	p, err := models.NewPolicyFromData(
		postgresql.AnyToUUID(values[2]),                       // id
		convert.AnyToString(values[1]),                        // language
		convert.AnyToString(values[5]),                        // rvvaID
		convert.AnyToString(values[6]),                        // uri
		bytes.NewBufferString(convert.AnyToString(values[8])), // content
	)
	if err != nil {
		return nil, err
	}

	return p.
		WithStatus(models.StatusFromString(convert.AnyToString(values[0]))). // Status
		WithTitle(convert.AnyToString(values[3])).                           // title
		WithDescription(convert.AnyToString(values[4])).                     // description
		WithTags(convert.AnyToStrings(values[7])...).                        // tags
		WithAudit(
			convert.AnyToDateTime(values[9]).UTC(),  // created
			convert.AnyToString(values[10]),         // createdBy
			convert.AnyToDateTime(values[11]).UTC(), // updated
			convert.AnyToString(values[12]),         // updatedBy
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
