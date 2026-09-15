package authorization

import (
	"encoding/base64"
	"log/slog"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func basicRequest(user string) *Request {
	uid := uuid.New()
	u, _ := url.Parse("https://openftv.nl/v1/policies")

	return &Request{
		UID:    &uid,
		URL:    u,
		Method: "GET",
		Headers: map[string][]string{
			"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(user+":secret"))},
		},
	}
}

// A basic-auth user carries no roles in the request; they live on the stored entity, in the
// shape the seed files use (a set) or as one string.
func TestIdentify_basicAuthRolesComeFromTheStoredEntity(t *testing.T) {
	t.Parallel()

	entities := map[string]any{"user::admin": []string{"admin", "user"}, "user::mickey": "author", "user::empty": nil}
	getter := func(uid string) (*models.Entity, uint64, error) {
		roles, ok := entities[uid]
		if !ok {
			return nil, 0, nil
		}

		attrs := models.NewAttributeSet()
		if roles != nil {
			attrs.AddAttributeKV(models.AttrRoles, roles)
		}

		return models.NewEntity("user", uid[len("user::"):], attrs), 1, nil
	}

	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), WithEntityGetter(getter), WithAuthenticator(authentication.NewDummy()))

	cases := []struct {
		user string
		want []string
	}{
		{"admin", []string{"admin", "user"}},
		{"mickey", []string{"author"}},
		{"empty", nil},
		{"unknown", nil},
	}

	for _, tc := range cases {
		t.Run(tc.user, func(t *testing.T) {
			t.Parallel()

			caller, err := a.Identify(basicRequest(tc.user))
			require.NoError(t, err)
			require.True(t, caller.IsAuthenticatedUser())
			assert.Equal(t, tc.user, caller.ID)
			assert.Equal(t, tc.want, caller.Roles)
		})
	}
}

func TestRolesOf(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"a", "b"}, rolesOf([]string{"a", "b"}))
	assert.Equal(t, []string{"a"}, rolesOf("a"))
	assert.Equal(t, []string{"a", "b"}, rolesOf([]any{"a", 1, "", "b"}))
	assert.Nil(t, rolesOf(""))
	assert.Nil(t, rolesOf(nil))
	assert.Nil(t, rolesOf(42))
}
