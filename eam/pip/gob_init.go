package pip

import (
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(time.Time{})
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(map[string]*attribute{})
}
