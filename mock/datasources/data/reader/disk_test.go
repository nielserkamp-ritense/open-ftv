package reader

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/dataspaces/rdw"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/dataspaces/rvig"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/dataspaces/vlierdam"
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

func TestLoadFromPath_Empty(t *testing.T) {
	t.Parallel()

	err := LoadFromPath(memory.New(nil), "")
	require.ErrorContains(t, err, "no data path given")
}

func TestLoadFromFS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		fsys    fs.FS
		table   string
		wantErr bool
	}{
		{name: "no metadata", fsys: os.DirFS("../../../../testdata/unittest/data/no-meta"), wantErr: true},
		{name: "bad metadata", fsys: os.DirFS("../../../../testdata/unittest/data/bad-meta"), wantErr: true},
		{
			name:    "missing datasource",
			fsys:    fstest.MapFS{"metadata.yaml": {Data: []byte("datasource:\n  - brv/schema.yaml\n")}},
			wantErr: true,
		},
		{name: "embedded rdw", fsys: rdw.FS, table: "brv.kenteken"},
		{name: "embedded rvig", fsys: rvig.FS, table: "brp.persoon"},
		{name: "embedded vlierdam", fsys: vlierdam.FS, table: "vlierdam.laadpaal"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := memory.New(nil)
			require.NotNil(t, s)

			err := LoadFromFS(s, tc.fsys)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, s.TableExists(tc.table))
			}
		})
	}
}

func TestFixPath(t *testing.T) {
	t.Parallel()

	p := &pathLoader{basePath: "tmp/data"}

	testCases := []struct {
		name string
		path string
		want string
	}{
		{name: "relative", path: "brv/schema.yaml", want: "tmp/data/brv/schema.yaml"},
		{name: "relative outside the base path", path: "../shared/klasse.csv", want: "tmp/shared/klasse.csv"},
		{name: "absolute starts at the root", path: "/opt/dataspace/../brv/schema.yaml", want: "opt/brv/schema.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, p.fixPath(tc.path))
		})
	}
}

func TestLoadMeta_Error(t *testing.T) {
	t.Parallel()

	t.Run("load meta - error", func(t *testing.T) {
		t.Parallel()

		p := &pathLoader{fsys: fstest.MapFS{}}
		err := p.loadMeta("this/is/not/a/valid/file")
		require.Error(t, err)
	})
}
