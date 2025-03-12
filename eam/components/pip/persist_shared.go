package pip

import (
	"context"
	"fmt"
	"sync"

	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"
)

// PathSeparator is the standard separator character to use with multi-level keys.
const PathSeparator = "/"

type wrapper struct {
	ctx      context.Context
	client   store.Store
	basePath string
	mutex    sync.RWMutex
}

func (w *wrapper) makeKey(key string) string {
	return fmt.Sprintf("%s%s", w.basePath, key)
}

func (w *wrapper) bugFix(in string) string {
	// the Valkeyrie/etcdv3 implementation sometimes removes a leading slash character from the key.
	if _, ok := w.client.(*etcdv3.Store); ok {
		return "/" + in
	}
	return in
}

var (
	readOptions  = &store.ReadOptions{}
	writeOptions = &store.WriteOptions{}
)
