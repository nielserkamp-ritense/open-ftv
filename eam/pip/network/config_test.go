package network

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_BadPath(t *testing.T) {
	t.Parallel()

	t.Run("bad path", func(t *testing.T) {
		cfg, err := LoadConfig("/this/is/not/a/valid/file/path")
		require.Error(t, err)
		require.Nil(t, cfg)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	d := t.TempDir()

	testCases := []struct {
		name    string
		file    string
		data    string
		wantErr bool
	}{
		{name: "invalid mime", file: "file0.txt", data: bad, wantErr: true},
		{name: "good yaml", file: "file1.yaml", data: yaml1},
		{name: "bad yaml", file: "file2.yaml", data: bad, wantErr: true},
		{name: "good json", file: "file1.json", data: json1},
		{name: "bad json", file: "file2.json", data: bad, wantErr: true},
		{name: "good toml", file: "file1.toml", data: toml1},
		{name: "bad toml", file: "file2.toml", data: bad, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(d, tc.file)

			f, err := os.Create(path)
			require.NoError(t, err)

			_, err = f.Write([]byte(tc.data))
			require.NoError(t, err)

			err = f.Close()
			require.NoError(t, err)

			cfg, err2 := LoadConfig(path)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, cfg)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, cfg)
			}
		})
	}
}

const (
	yaml1 = `
sources:
  - name: remote1
    description: first remote PIP
    requests:
      - name: attributes
        description: get all attributes
        uri: http://localhost:8080/v1/attributes
        timeout: 30s
        interval: 5m
      - name: entities
        description: get all entities
        uri: http://localhost:8080/v1/entities
        timeout: 30s
        interval: 15m
`

	json1 = `{
 "sources": [
  {"name": "remote1","requests":[
   {"name": "attributes","description":"get all attributes","uri":"http://localhost:8080/v1/attributes"}]
  }
 ]
}`

	toml1 = `
sources = [
 { name = "remote2", description = "second remote PIP" },
 { name = "remote1", description = "first remote PIP", requests = [
   { name = "attributes", description = "get all attributes", uri = "http://localhost:8080/v1/attributes" },
   { name = "entities", description = "get all entities", uri = "http://localhost:8080/v1/entities" }
  ] }
]`

	bad = `oops`
)
