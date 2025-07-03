package reader

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory"
)

func TestLoadFromPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		schema  *schema.Dataspace
		path    string
		wantErr bool
	}{
		{name: "invalid path", path: "/this/is/never/valid/you/know/it", wantErr: true},
		{name: "no metadata", path: "../../../../testdata/unittest/data/no-meta", wantErr: true},
		{name: "bad metadata", path: "../../../../testdata/unittest/data/bad-meta", wantErr: true},
		{name: "invalid dataspace path", path: "../../../../testdata/unittest/data/invalid-dataspace", wantErr: true},
		{name: "bad dataspace yaml", path: "../../../../testdata/unittest/data/bad-dataspace-yaml", wantErr: true},
		{name: "bad dataspace json", path: "../../../../testdata/unittest/data/bad-dataspace-json", wantErr: true},
		{name: "bad dataspace unknown", path: "../../../../testdata/unittest/data/bad-dataspace-unknown", wantErr: true},
		{name: "invalid datasource path", path: "../../../../testdata/unittest/data/invalid-datasource", wantErr: true},
		{name: "bad datasource yaml", path: "../../../../testdata/unittest/data/bad-datasource-yaml", wantErr: true},
		{name: "bad datasource json", path: "../../../../testdata/unittest/data/bad-datasource-json", wantErr: true},
		{name: "bad datasource unknown", path: "../../../../testdata/unittest/data/bad-datasource-unknown", wantErr: true},
		{name: "invalid data path", path: "../../../../testdata/unittest/data/invalid-data", wantErr: true},
		{name: "bad data yaml", path: "../../../../testdata/unittest/data/bad-data-yaml", wantErr: true},
		{name: "bad data json", path: "../../../../testdata/unittest/data/bad-data-json", wantErr: true},
		{name: "bad data csv", path: "../../../../testdata/unittest/data/bad-data-csv", wantErr: true},
		{name: "bad data unknown", path: "../../../../testdata/unittest/data/bad-data-unknown", wantErr: true},
		{name: "all good", path: "../../../../testdata/dataspaces/fds"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := memory.New(tc.schema)
			require.NotNil(t, s)

			err := LoadFromPath(s, tc.path)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadMeta_Error(t *testing.T) {
	t.Parallel()

	t.Run("load meta - error", func(t *testing.T) {
		t.Parallel()

		p := &pathLoader{}
		err := p.loadMeta("/this/is/not/a/valid/file")
		require.Error(t, err)
	})
}
