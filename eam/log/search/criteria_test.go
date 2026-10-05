package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCriteria_Test_duplicateBundle(t *testing.T) {
	t.Parallel()

	// no request types are given: the message must name the bundle, not index that list.
	c := &Criteria{Recent: time.Hour, Bundles: []int64{3, 3}}

	require.ErrorContains(t, c.Test(), "duplicate bundle: 3")
}
