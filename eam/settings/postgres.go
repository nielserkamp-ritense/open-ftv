package settings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// settingsColumnCount is the number of columns selected in the settings query.
const settingsColumnCount = 7

// NewSettingsDBWithPool instantiates a PostgreSQL-backed settings store using the given connection pool.
func NewSettingsDBWithPool(pool *postgresql.Postgres) *SettingsDB {
	return &SettingsDB{p: pool, now: time.Now}
}

// SettingsDB wraps a PostgreSQL connection pool with settings management functions.
type SettingsDB struct {
	p   *postgresql.Postgres
	now func() time.Time // for time-sensitive unit-tests.
}

// GetSettings returns the manager ui settings, or the defaults if none have been saved yet.
func (db *SettingsDB) GetSettings(ctx context.Context) (*oas.Settings, error) {
	sql := `SELECT header_title,header_color,title_color,logo,logo_media_type,updated,updated_by FROM settings WHERE singleton_id = TRUE`

	var (
		s   *oas.Settings
		err error
	)

	err2 := db.p.Query(ctx, sql, nil, func(values []any) bool {
		s, err = settingsFromValues(values)
		return false // only one possible row.
	})

	if err != nil || err2 != nil {
		return nil, errors.Join(err, err2)
	}

	if s == nil {
		// OpenAPI: when nothing has been saved yet, return the built-in defaults.
		return Defaults(), nil
	}

	return s, nil
}

// UpdateSettings saves the manager ui settings, replacing any previous value.
func (db *SettingsDB) UpdateSettings(ctx context.Context, user identity.Principal, in *oas.Settings) (*oas.Settings, error) {
	now := db.now().UTC()

	sql := `INSERT INTO settings (singleton_id,header_title,header_color,title_color,logo,logo_media_type,created,created_by,updated,updated_by) VALUES (TRUE,$1,$2,$3,$4,$5,$6,$7,$6,$7) ON CONFLICT (singleton_id) DO UPDATE SET header_title=$1,header_color=$2,title_color=$3,logo=$4,logo_media_type=$5,updated=$6,updated_by=$7`

	var logoMediaType *string

	if in.LogoMediaType != "" {
		s := string(in.LogoMediaType)
		logoMediaType = &s
	}

	params := []any{in.HeaderTitle, in.HeaderColor, in.TitleColor, in.Logo, logoMediaType, now, user.ID}

	if _, err := db.p.Exec(identity.WithContext(ctx, user), sql, params); err != nil {
		return nil, err
	}

	out := *in
	out.Updated = now.Format(time.RFC3339Nano)
	out.UpdatedBy = optionalPrincipal(user.ID)

	return &out, nil
}

func settingsFromValues(values []any) (*oas.Settings, error) {
	if len(values) != settingsColumnCount {
		return nil, fmt.Errorf("invalid number of values")
	}

	var updated string
	if d := convert.AnyToDateTime(values[5]); !d.IsZero() {
		updated = d.Format(time.RFC3339Nano)
	}

	logo, _ := values[3].([]byte)

	return &oas.Settings{
		HeaderTitle:   convert.AnyToString(values[0]),
		HeaderColor:   convert.AnyToString(values[1]),
		TitleColor:    convert.AnyToString(values[2]),
		Logo:          logo,
		LogoMediaType: oas.SettingsLogoMediaType(convert.AnyToString(values[4])),
		Updated:       updated,
		UpdatedBy:     optionalPrincipal(convert.AnyToString(values[6])),
	}, nil
}

// optionalPrincipal returns nil for an absent attribution, so the field is omitted from the
// response rather than serialized as a principal identifying nobody.
func optionalPrincipal(id string) *oas.Principal {
	if id == "" {
		return nil
	}

	return &oas.Principal{Id: id}
}
