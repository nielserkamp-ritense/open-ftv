package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin_Fix(t *testing.T) {
	t.Parallel()

	t1 := &Table{Object: Object{Parent: Parent{ID: "t1"}}}
	t2 := &Table{Object: Object{Parent: Parent{ID: "t2"}}}
	t3 := &Table{Object: Object{Parent: Parent{ID: "t3"}}}

	ds := &Datasource{
		Parent:      Parent{ID: "ds1"},
		Description: "source1",
		Tables:      []*Table{t1, t2, t3},
	}

	ds.Fix(nil)

	testCases := []struct {
		name        string
		j           *Join
		wantPrimary *Table
		wantTable   *Table
		wantJoinID  string
	}{
		{
			name:        "without join ID",
			j:           &Join{Target: "t1", Source: "t2"},
			wantPrimary: t1,
			wantTable:   t2,
			wantJoinID:  "t2",
		},
		{
			name:        "with join ID",
			j:           &Join{Target: "t2", Source: "t3", JoinID: "abc"},
			wantPrimary: t2,
			wantTable:   t3,
			wantJoinID:  "abc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.j.Fix(ds)

			assert.Equal(t, tc.wantPrimary, tc.j.GetTarget())
			assert.Equal(t, tc.wantTable, tc.j.GetSource())
			assert.Equal(t, tc.wantJoinID, tc.j.GetJoinID())
		})
	}
}
