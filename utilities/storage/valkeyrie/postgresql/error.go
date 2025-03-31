package postgresql

import "fmt"

func dsnError(err error) error {
	return fmt.Errorf("dsn parsing failed: %w", err)
}

func (p *pool) baseError(cmd string) (string, []any) {
	return "%s:%d/%s: %s failed: ", []any{p.cfg.ConnConfig.Config.Host, p.cfg.ConnConfig.Config.Port, p.cfg.ConnConfig.Config.Database, cmd}
}

func (p *pool) connectionError(err error) error {
	s, params := p.baseError("connection")
	return fmt.Errorf(s+" %w", append(params, err)...)
}

func (p *pool) pingError(err error) error {
	s, params := p.baseError("ping")
	return fmt.Errorf(s+" %w", append(params, err)...)
}

func (p *pool) txError(err error) error {
	s, params := p.baseError("begin transaction")
	return fmt.Errorf(s+" %w", append(params, err)...)
}

func (p *pool) queryError(err error, q string, params ...any) error {
	s, params2 := p.baseError("query")
	return fmt.Errorf(s+" '%s' %#v: %w", append(params2, q, params, err)...)
}

func (p *pool) scanError(err error, q string, params ...any) error {
	s, params2 := p.baseError("row scan")
	return fmt.Errorf(s+" '%s' %#v: %w", append(params2, q, params, err)...)
}

func (db *pgDB) txError(err error) error {
	if p, ok := db.pool.(*pool); ok {
		return p.txError(err)
	}
	return fmt.Errorf("begin transaction failed: %w", err)
}

func (db *pgDB) queryError(err error, q string, params ...any) error {
	if p, ok := db.pool.(*pool); ok {
		return p.queryError(err, q, params...)
	}
	return fmt.Errorf("query failed: '%s' %#v: %w", q, params, err)
}

func (db *pgDB) scanError(err error, q string, params ...any) error {
	if p, ok := db.pool.(*pool); ok {
		return p.scanError(err, q, params...)
	}
	return fmt.Errorf("row scan failed: '%s' %#v: %w", q, params, err)
}
