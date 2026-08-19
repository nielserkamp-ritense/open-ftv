package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestTransformation_Fix(t *testing.T) {
	t.Parallel()

	f1 := &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: enums.IntegerType}
	f2 := &Field{Object: Object{Parent: Parent{ID: "f2"}}, Type: enums.StringType}
	f3 := &Field{Object: Object{Parent: Parent{ID: "f3"}}, Type: enums.FloatType}
	f4 := &Field{Object: Object{Parent: Parent{ID: "f4"}}, Type: enums.DateType}

	age := &Transformation{
		Object:             Object{Parent: Parent{ID: "age"}},
		TransformationType: enums.TransformAge,
		ResultType:         enums.IntegerType,
		IsPII:              true,
		InputFields:        map[int]string{1: "f4"},
	}

	t1 := &Table{
		Object: Object{
			Parent: Parent{ID: "t1"},
			Fields: []*Field{f1, f2, f3, f4},
		},
		Transforms: []*Transformation{age},
	}

	ds := &Datasource{
		Parent: Parent{ID: "ds1"},
		Tables: []*Table{t1},
	}
	ds.Fix(nil)

	testCases := []struct {
		name     string
		tr       *Transformation
		t        *Table
		wantFQID string
	}{
		{
			name: "simple compare",
			tr: &Transformation{
				Object:             Object{Parent: Parent{ID: "tr1"}},
				TransformationType: enums.TransformCompare,
				ResultType:         enums.BooleanType,
				CompareType:        enums.IsGreaterOrEqual,
				InputFields:        map[int]string{1: "f1"},
				InputValues:        map[int]any{2: 18.0},
			},
			t:        t1,
			wantFQID: "t1.tr1",
		},
		{
			name: "case-insensitive compare - output as integer",
			tr: &Transformation{
				Object:             Object{Parent: Parent{ID: "tr2"}},
				TransformationType: enums.TransformCompare,
				ResultType:         enums.IntegerType,
				CompareType:        enums.IsEqual,
				Insensitive:        true,
				InputFields:        map[int]string{1: "f2"},
				InputValues:        map[int]any{2: "harry potter"},
			},
			t:        t1,
			wantFQID: "t1.tr2",
		},
		{
			name: "regex",
			tr: &Transformation{
				Object:             Object{Parent: Parent{ID: "tr3"}},
				TransformationType: enums.TransformCompare,
				ResultType:         enums.BooleanType,
				CompareType:        enums.MatchRegex,
				InputFields:        map[int]string{1: "f2"},
				InputValues:        map[int]any{2: "[H|h]ello .*"},
			},
			t:        t1,
			wantFQID: "t1.tr3",
		},
		{
			name: "recursive transform - output as string",
			tr: &Transformation{
				Object:             Object{Parent: Parent{ID: "tr4"}},
				TransformationType: enums.TransformCompare,
				ResultType:         enums.StringType,
				CompareType:        enums.IsGreaterOrEqual,
				InputTransforms:    map[int]string{1: "age"},
				InputValues:        map[int]any{2: 18.0},
			},
			t:        t1,
			wantFQID: "t1.tr4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tr := tc.tr
			tr.Fix(tc.t)
			assert.Equal(t, tc.wantFQID, tr.FQID())
			assert.Len(t, tr.fields, len(tc.tr.InputFields))
			assert.Len(t, tr.transforms, len(tc.tr.InputTransforms))

			got, err := json.Marshal(tr)
			require.NoError(t, err)
			require.NotNil(t, got)

			tr2 := new(Transformation)
			err = tr2.UnmarshalJSON(got)
			require.NoError(t, err)

			assert.Equal(t, tr.ID, tr2.ID)
			assert.Equal(t, tr.Description, tr2.Description)
			assert.Equal(t, tr.TransformationType, tr2.TransformationType)
			assert.Equal(t, tr.ResultType, tr2.ResultType)
			assert.Equal(t, tr.CompareType, tr2.CompareType)
			assert.Equal(t, tr.IsPII, tr2.IsPII)
			assert.Equal(t, tr.InputFields, tr2.InputFields)
			assert.Equal(t, tr.InputTransforms, tr2.InputTransforms)
			assert.Equal(t, tr.InputValues, tr2.InputValues)

			got, err = yaml.Marshal(tr)
			require.NoError(t, err)
			require.NotNil(t, got)

			tr2 = new(Transformation)
			err = tr2.UnmarshalYAML(got)
			require.NoError(t, err)

			assert.Equal(t, tr.ID, tr2.ID)
			assert.Equal(t, tr.Description, tr2.Description)
			assert.Equal(t, tr.TransformationType, tr2.TransformationType)
			assert.Equal(t, tr.ResultType, tr2.ResultType)
			assert.Equal(t, tr.CompareType, tr2.CompareType)
			assert.Equal(t, tr.IsPII, tr2.IsPII)
			assert.Equal(t, tr.InputFields, tr2.InputFields)
			assert.Equal(t, tr.InputTransforms, tr2.InputTransforms)
			assert.Equal(t, tr.InputValues, tr2.InputValues)
		})
	}
}

func TestTransformation_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("transformation unmarshal json error", func(t *testing.T) {
		t.Parallel()

		tr := new(Transformation)
		err := tr.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestTransformation_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	t.Run("transformation unmarshal yaml error", func(t *testing.T) {
		t.Parallel()

		tr := new(Transformation)
		err := tr.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
	})
}
