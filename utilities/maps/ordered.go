package maps

import (
	"cmp"
	"slices"
)

// ProcessOrdered process a map in ordered sequence of its keys.
func ProcessOrdered[K cmp.Ordered, V any](m map[K]V, f func(k K, v V)) {
	l := len(m)
	if l == 0 {
		return
	}

	keys := make([]K, 0, l)
	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for i := range keys {
		k := keys[i]
		f(k, m[k])
	}
}
