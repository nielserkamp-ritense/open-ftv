package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRows_RemoveFields(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()

	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f1": 2, "f2": "goodbye mars", "f3": false, "f4": 12.345}, def: def1}
	r3 := &Row{Data: map[string]any{"f1": 3, "f2": "hello world", "f3": true, "f4": 123.45}, def: def1}
	r4 := &Row{Data: map[string]any{"f1": 4, "f2": "goodbye mars", "f3": false, "f4": 1234.5}, def: def1}

	testCases := []struct {
		Name   string
		in     Rows
		fields []string
		want   Rows
	}{
		{
			Name:   "empty",
			in:     Rows{},
			fields: []string{"f1", "f2", "f3"},
			want:   Rows{},
		},
		{
			Name:   "no matches",
			in:     Rows{r1, r2, r3, r4},
			fields: []string{"f6", "f8", "f5"},
			want:   Rows{r1, r2, r3, r4},
		},
		{
			Name:   "single match",
			in:     Rows{r1, r2, r3, r4},
			fields: []string{"f6", "f3", "f5"},
			want: Rows{
				{Data: map[string]any{"f1": 1, "f2": "hello world", "f4": 1.2345}, def: def1},
				{Data: map[string]any{"f1": 2, "f2": "goodbye mars", "f4": 12.345}, def: def1},
				{Data: map[string]any{"f1": 3, "f2": "hello world", "f4": 123.45}, def: def1},
				{Data: map[string]any{"f1": 4, "f2": "goodbye mars", "f4": 1234.5}, def: def1},
			},
		},
		{
			Name:   "few matches",
			in:     Rows{r1, r2, r3, r4},
			fields: []string{"f6", "f4", "f2"},
			want: Rows{
				{Data: map[string]any{"f1": 1, "f3": true}, def: def1},
				{Data: map[string]any{"f1": 2, "f3": false}, def: def1},
				{Data: map[string]any{"f1": 3, "f3": true}, def: def1},
				{Data: map[string]any{"f1": 4, "f3": false}, def: def1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.RemoveFields(tc.fields)
			require.Equal(t, len(tc.want), len(got))

			for i := range tc.want {
				assert.Equal(t, tc.want[i].def, got[i].def)
				assert.EqualValues(t, tc.want[i].Data, got[i].Data)
			}
		})
	}
}

func TestRows_MarshalCSV(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()

	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f1": 2, "f2": "goodbye mars", "f3": false, "f4": 12.345}, def: def1}
	r3 := &Row{Data: map[string]any{"f1": 3, "f2": "hello world", "f3": true, "f4": 123.45}, def: def1}

	testCases := []struct {
		name string
		in   Rows
		want string
	}{
		{
			name: "no data",
			in:   Rows{},
		},
		{
			name: "single record",
			in:   Rows{r3},
			want: `"f1","f2","f3","f4"
3,"hello world","true",123.45
`,
		},
		{
			name: "few records",
			in:   Rows{r3, r2, r1},
			want: `"f1","f2","f3","f4"
3,"hello world","true",123.45
2,"goodbye mars","false",12.345
1,"hello world","true",1.2345
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.in.MarshalCSV()
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(got))
		})
	}
}

func TestRows_Encode(t *testing.T) {
	t.Parallel()

	t.Run("encode", func(t *testing.T) {
		t.Parallel()

		def1 := makeRowDatasource1()

		r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
		r2 := &Row{Data: map[string]any{"f1": 2, "f2": "goodbye mars", "f3": false, "f4": 12.345}, def: def1}
		r3 := &Row{Data: map[string]any{"f1": 3, "f2": "hello world", "f3": true, "f4": 123.45}, def: def1}

		r := Rows{r3, r2, r1}

		got := r.Encode()
		want := []any{
			map[string]any{"f1": 3, "f2": "hello world", "f3": true, "f4": 123.45},
			map[string]any{"f1": 2, "f2": "goodbye mars", "f3": false, "f4": 12.345},
			map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345},
		}

		assert.EqualValues(t, want, got)
	})
}
