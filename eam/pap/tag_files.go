package pap

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// ReadTagFiles reads the policy tags from all YAML and JSON files under path, recursing into subdirectories.
// An empty path means no tags. On error, it returns the tags read so far along with the error.
func ReadTagFiles(path string) ([]*policies.Tag, error) {
	if path == "" {
		return nil, nil
	}

	tags := make([]*policies.Tag, 0)
	root := path

	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// A mounted ConfigMap keeps its files in a ..<timestamp> directory, linked from the top.
			if path != root && strings.HasPrefix(d.Name(), "..") {
				return filepath.SkipDir
			}

			return nil
		}

		var list []*policies.Tag

		switch strings.ToLower(filepath.Ext(path)) {
		case ".yaml", ".yml":
			list, err = readTagFile(path, func(f *os.File, v any) error { return yaml.NewDecoder(f).Decode(v) })
		case ".json":
			list, err = readTagFile(path, func(f *os.File, v any) error { return json.NewDecoder(f).Decode(v) })
		default:
			return nil
		}

		if err != nil {
			return err
		}

		tags = append(tags, list...)

		return nil
	})

	return tags, err
}

func readTagFile(path string, decode func(*os.File, any) error) ([]*policies.Tag, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	list := make([]*policies.Tag, 0)
	if err := decode(f, &list); err != nil {
		return nil, err
	}

	return list, nil
}
