package pip

import "encoding/gob"

func init() {
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(map[string]*attribute{})
}
