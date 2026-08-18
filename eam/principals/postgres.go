package principals

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// recordColumnCount is the number of columns selected by the resolve query.
const recordColumnCount = 4

// NewDBWithPool instantiates a PostgreSQL-backed principal store using the given connection pool.
func NewDBWithPool(pool *postgresql.Postgres) *DB {
	return &DB{p: pool}
}

// DB wraps a PostgreSQL connection pool with principal storage.
type DB struct {
	p *postgresql.Postgres
}

// Upsert records that p was seen acting on the management plane.
//
// The conflict target is the primary key, not (issuer, subject): principal.id IS the attribution
// string, so a row backfilled by migration 00013 (which knows no issuer) and the same subject
// arriving with a real issuer are one row, not two. See docs/adr/0004.
//
// display_name, email and issuer are all COALESCEd so a request that carries less than a full JWT
// never erases what the manager already knew. A user principal can be resolved without any token
// at all (eam/pep.DeterminePrincipal reads a `user::<id>` attribute), which would otherwise blank
// the issuer recorded on an earlier request.
func (db *DB) Upsert(ctx context.Context, p *identity.Principal) error {
	sql := `INSERT INTO principal (id, kind, issuer, display_name, email, first_seen, last_seen)
 VALUES ($1,$2,$3,$4,$5,now(),now())
 ON CONFLICT (id) DO UPDATE SET
 kind=EXCLUDED.kind,
 issuer=COALESCE(NULLIF(EXCLUDED.issuer, ''), principal.issuer),
 display_name=COALESCE(EXCLUDED.display_name, principal.display_name),
 email=COALESCE(EXCLUDED.email, principal.email),
 last_seen=now()`

	params := []any{p.ID, KindUser, p.Issuer, nullable(p.Name), nullable(p.Email)}

	if _, err := db.p.Exec(ctx, sql, params); err != nil {
		return fmt.Errorf("failed to upsert principal: %w", err)
	}

	return nil
}

// Resolve returns a Record for each of the given ids that is known.
//
// It is a single batched query rather than one per id: a list view resolves every distinct
// author it displays in one round-trip.
func (db *DB) Resolve(ctx context.Context, ids []string) (map[string]Record, error) {
	out := make(map[string]Record, len(ids))

	if len(ids) == 0 {
		return out, nil
	}

	sql := `SELECT id, kind, display_name, email FROM principal WHERE id = ANY($1)`
	params := []any{dedupe(ids)}

	var err error

	err2 := db.p.Query(ctx, sql, params, func(values []any) bool {
		var r Record
		if r, err = recordFromValues(values); err != nil {
			return false
		}

		out[r.ID] = r

		return true
	})

	if err != nil || err2 != nil {
		return nil, fmt.Errorf("failed to resolve principals: %w", errors.Join(err, err2))
	}

	return out, nil
}

func recordFromValues(values []any) (Record, error) {
	if len(values) != recordColumnCount {
		return Record{}, fmt.Errorf("invalid number of values")
	}

	return Record{
		ID:    convert.AnyToString(values[0]),
		Kind:  convert.AnyToString(values[1]),
		Name:  convert.AnyToString(values[2]),
		Email: convert.AnyToString(values[3]),
	}, nil
}

// nullable maps an empty string to a SQL NULL, so COALESCE in the upsert can tell "the IdP did
// not tell us" from "the IdP told us it is empty".
func nullable(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// dedupe returns ids without duplicates or empty values, so a list view of 100 objects written
// by two people queries for two ids.
func dedupe(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}
