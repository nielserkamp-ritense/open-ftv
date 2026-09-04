package pip

import (
	"fmt"
	"sync"

	"github.com/kvtools/valkeyrie/store"
)

// PathSeparator is the standard separator character to use with multi-level keys.
const PathSeparator = "/"

type kvWrapper struct {
	client   store.Store
	basePath string
	mutex    sync.RWMutex
}

func (w *kvWrapper) makeKey(key string) string {
	return fmt.Sprintf("%s%s", w.basePath, key)
}

var (
	readOptions  = &store.ReadOptions{}
	writeOptions = &store.WriteOptions{}
)
