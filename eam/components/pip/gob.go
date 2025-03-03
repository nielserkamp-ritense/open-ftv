package pip

import "encoding/gob"

func init() {
	gob.Register(map[string]any{})
}
