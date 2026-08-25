package settings

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func newMockPostgres(t *testing.T, ctx context.Context, now time.Time) (pgxmock.PgxPoolIface, *SettingsDB) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	require.NotNil(t, mock)

	pool, err2 := postgresql.NewWithPool(ctx, "postgres://localhost:5432/table", time.Minute, 3, mock)
	require.NoError(t, err2)
	require.NotNil(t, pool)

	db := NewSettingsDBWithPool(pool)
	db.now = func() time.Time { return now }

	return mock, db
}

func TestPostgres_GetSettings_NoRow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, db := newMockPostgres(t, ctx, time.Now())

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT header_title,header_color,title_color,logo,logo_media_type,updated,updated_by FROM settings WHERE singleton_id = TRUE").
		WillReturnRows(pgxmock.NewRows([]string{"header_title", "header_color", "title_color", "logo", "logo_media_type", "updated", "updated_by"}))
	mock.ExpectCommit()

	got, err := db.GetSettings(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "OpenFTV beheeromgeving", got.HeaderTitle)
	assert.Equal(t, "#F7E8E8", got.HeaderColor)
	assert.Equal(t, "#000000", got.TitleColor)
	assert.Empty(t, got.Logo)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgres_GetSettings_Existing(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, db := newMockPostgres(t, ctx, time.Now())

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT header_title,header_color,title_color,logo,logo_media_type,updated,updated_by FROM settings WHERE singleton_id = TRUE").
		WillReturnRows(pgxmock.NewRows([]string{"header_title", "header_color", "title_color", "logo", "logo_media_type", "updated", "updated_by"}).
			AddRow("ACME", "#112233", "#334455", []byte{1, 2, 3}, "image/png", "2025-08-15T07:53:41.493415Z", "alice@wonderland.cc"))
	mock.ExpectCommit()

	got, err := db.GetSettings(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "ACME", got.HeaderTitle)
	assert.Equal(t, "#112233", got.HeaderColor)
	assert.Equal(t, "#334455", got.TitleColor)
	assert.Equal(t, []byte{1, 2, 3}, got.Logo)
	assert.Equal(t, oas.SettingsLogoMediaType("image/png"), got.LogoMediaType)
	assert.Equal(t, "alice@wonderland.cc", got.UpdatedBy.Id)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgres_UpdateSettings(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Date(2025, 8, 15, 7, 53, 41, 0, time.UTC)
	mock, db := newMockPostgres(t, ctx, now)

	in := &oas.Settings{HeaderTitle: "ACME", HeaderColor: "#112233", TitleColor: "#334455"}
	user := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "alice@wonderland.cc"}

	mock.ExpectBegin()
	mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").WithArgs(user.ID).WillReturnResult(pgxmock.NewResult("SELECT", 1))
	mock.ExpectExec("INSERT INTO settings (singleton_id,header_title,header_color,title_color,logo,logo_media_type,created,created_by,updated,updated_by) VALUES (TRUE,$1,$2,$3,$4,$5,$6,$7,$6,$7) ON CONFLICT (singleton_id) DO UPDATE SET header_title=$1,header_color=$2,title_color=$3,logo=$4,logo_media_type=$5,updated=$6,updated_by=$7").
		WithArgs("ACME", "#112233", "#334455", []byte(nil), (*string)(nil), now, user.ID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	got, err := db.UpdateSettings(ctx, user, in)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "ACME", got.HeaderTitle)
	assert.Equal(t, user.ID, got.UpdatedBy.Id)
	assert.Equal(t, now.Format(time.RFC3339Nano), got.Updated)

	require.NoError(t, mock.ExpectationsWereMet())
}
