package openfga

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty"},
		{name: "no escape", in: "this doesn't escape\nat all!", want: "this doesn't escape\nat all!"},
		{name: "escape colon", in: "x:y", want: "eDp5"},
		{name: "escape at-sign", in: "x@y", want: "eEB5"},
		{name: "escape hash", in: "x#y", want: "eCN5"},
		{name: "escape all", in: "x:y#z@q", want: "eDp5I3pAcQ=="},
		{name: "escape uri", in: "http://domain.io/some/path#mark?x=y", want: "aHR0cDovL2RvbWFpbi5pby9zb21lL3BhdGgjbWFyaz94PXk="},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalize(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReadTuples(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		wantErr bool
		want    basicTuples
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "bad json",
			data:    "not really json, is it?",
			wantErr: true,
		},
		{
			name:    "not a json slice",
			data:    `{"subject":{"type":"x","id":"y"},"predicate":"q","object":{"type":"a","id":"b"}}`,
			wantErr: true,
		},
		{
			name: "single entry",
			data: `[{"subject":{"type":"x","id":"y"},"predicate":"q","object":{"type":"a","id":"b"}}]`,
			want: basicTuples{
				basicTuple{Subject: basicEntity{Type: "x", ID: "y"}, Predicate: "q", Object: basicEntity{Type: "a", ID: "b"}},
			},
		},
		{
			name: "multiple entries",
			data: `[
{"subject":{"type":"x1","id":"y1"},"predicate":"q","object":{"type":"a3","id":"b3"}},
{"subject":{"type":"x2","id":"y2"},"predicate":"p","object":{"type":"a2","id":"b2"}},
{"subject":{"type":"x3","id":"y3"},"predicate":"q","object":{"type":"a1","id":"b1"}}
]`,
			want: basicTuples{
				basicTuple{Subject: basicEntity{Type: "x1", ID: "y1"}, Predicate: "q", Object: basicEntity{Type: "a3", ID: "b3"}},
				basicTuple{Subject: basicEntity{Type: "x2", ID: "y2"}, Predicate: "p", Object: basicEntity{Type: "a2", ID: "b2"}},
				basicTuple{Subject: basicEntity{Type: "x3", ID: "y3"}, Predicate: "q", Object: basicEntity{Type: "a1", ID: "b1"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readTuples(strings.NewReader(tc.data))
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				for i := range got {
					assert.Equal(t, tc.want[i].Subject.key(), got[i].User)
					assert.Equal(t, tc.want[i].Predicate, got[i].Relation)
					assert.Equal(t, tc.want[i].Object.key(), got[i].Object)
				}
			}
		})
	}
}

func TestReadTuples_BadReader(t *testing.T) {
	t.Run("bad reader", func(t *testing.T) {
		f, err := os.Open("../../../../testdata/policies/openfga/doelbinding.relations")
		require.NoError(t, err)
		f.Close()

		_, err = readTuples(f)
		require.Error(t, err)
	})

}
