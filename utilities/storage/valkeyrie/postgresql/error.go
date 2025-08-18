package postgresql

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

func (db *pgDB) txError(err error) error {
	if p, ok := db.pool.(*pool.Pool); ok {
		return p.TxError(err)
	}
	return fmt.Errorf("begin transaction failed: %w", err)
}

func (db *pgDB) queryError(err error, q string, params ...any) error {
	if p, ok := db.pool.(*pool.Pool); ok {
		return p.QueryError(err, q, params...)
	}
	return fmt.Errorf("query failed: '%s' %#v: %w", q, params, err)
}

func (db *pgDB) scanError(err error, q string, params ...any) error {
	if p, ok := db.pool.(*pool.Pool); ok {
		return p.ScanError(err, q, params...)
	}
	return fmt.Errorf("row scan failed: '%s' %#v: %w", q, params, err)
}
