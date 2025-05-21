package schema

import (
	"bytes"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestField_Fix(t *testing.T) {
	t.Parallel()

	persoon := &Table{
		Object: Object{Parent: Parent{ID: "persoon"}, Description: "Persoonsgegevens"},
	}

	geboorte := &Field{
		Object: Object{
			Parent:      Parent{ID: "geboorte", parent: &persoon.Parent},
			Description: "Geboorte gegevens",
			parentTable: persoon,
		},
	}

	quiz := &Field{
		Object: Object{
			Parent:      Parent{ID: "quiz1", parent: &persoon.Parent},
			Description: "Quiz 1",
			parentTable: persoon,
		},
	}

	testCases := []struct {
		name        string
		field       *Field
		parent      *Object
		parentTable *Table
		parentField *Field
		wantFQDN    string
	}{
		{
			name: "simple table field",
			field: &Field{
				Object: Object{Parent: Parent{ID: "achternaam"}, Description: "Achternaam"},
				Type:   types.StringType,
				IsPII:  true,
			},
			parent:      &persoon.Object,
			parentTable: persoon,
			wantFQDN:    "persoon.achternaam",
		},
		{
			name: "simple sub-field",
			field: &Field{
				Object: Object{Parent: Parent{ID: "jaar"}, Description: "Geboorte jaar"},
				Type:   types.UnsignedIntegerType,
			},
			parent:      &geboorte.Object,
			parentField: geboorte,
			wantFQDN:    "persoon.geboorte.jaar",
		},
		{
			name: "structured table field",
			field: &Field{
				Object: Object{
					Parent:      Parent{ID: "voorkeuren"},
					Description: "Persoonlijke voorkeuren",
					Fields: []*Field{
						{Object: Object{Parent: Parent{ID: "kleur"}, Description: "Kleur"}, Type: types.StringType},
						{Object: Object{Parent: Parent{ID: "muziek"}, Description: "Muziek"}, Type: types.StringType},
						{Object: Object{Parent: Parent{ID: "boek"}, Description: "Boek"}, Type: types.StringType},
					},
				},
				IsPII: true,
			},
			parent:      &persoon.Object,
			parentTable: persoon,
			wantFQDN:    "persoon.voorkeuren",
		},
		{
			name: "structured sub-field",
			field: &Field{
				Object: Object{
					Parent:      Parent{ID: "voorkeuren"},
					Description: "Persoonlijke voorkeuren",
					Fields: []*Field{
						{Object: Object{Parent: Parent{ID: "kleur"}, Description: "Kleur"}, Type: types.StringType},
						{Object: Object{Parent: Parent{ID: "muziek"}, Description: "Muziek"}, Type: types.StringType},
						{Object: Object{Parent: Parent{ID: "boek"}, Description: "Boek"}, Type: types.StringType},
					},
				},
				IsPII: true,
			},
			parent:      &quiz.Object,
			parentField: quiz,
			wantFQDN:    "persoon.quiz1.voorkeuren",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.field
			f.Fix(tc.parent, tc.parentTable, tc.parentField)

			assert.Equal(t, &tc.parent.Parent, f.parent)
			assert.Equal(t, tc.parentTable, f.parentTable)
			assert.Equal(t, tc.parentField, f.parentField)
			assert.Equal(t, tc.wantFQDN, f.FQDN())

			for _, f2 := range f.Fields {
				assert.Equal(t, &f.Parent, f2.parent)
				assert.Nil(t, f2.parentTable)
				assert.Equal(t, f, f2.parentField)
			}

			d, err := json.Marshal(f)
			require.NoError(t, err)
			require.NotNil(t, d)

			var f3 *Field
			err = json.Unmarshal(d, &f3)
			require.NoError(t, err)
			assert.Equal(t, f.ID, f3.ID)
			assert.Equal(t, f.Description, f3.Description)

			d, err = yaml.Marshal(f)
			require.NoError(t, err)
			require.NotNil(t, d)

			f4 := new(Field)
			err = yaml.Unmarshal(d, f4)
			require.NoError(t, err)
			assert.Equal(t, f.ID, f4.ID)
			assert.Equal(t, f.Description, f4.Description)
		})
	}
}

func TestField_JSON_error(t *testing.T) {
	t.Parallel()

	t.Run("field json error", func(t *testing.T) {
		t.Parallel()

		b := bytes.NewBufferString("\000\001")

		f := new(Field)
		err := f.UnmarshalJSON(b.Bytes())
		require.Error(t, err)
	})
}

func TestField_YAML_error(t *testing.T) {
	t.Parallel()

	t.Run("field yaml error", func(t *testing.T) {
		t.Parallel()

		b := bytes.NewBufferString("\000\001")

		f := new(Field)
		err := f.UnmarshalYAML(b.Bytes())
		require.Error(t, err)
	})
}

func TestField_ConvertValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		f    *Field
		v    any
		want any
	}{
		{name: "nil", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.StringType, IsArray: true}},
		{name: "string", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.StringType}, v: 123456, want: "123456"},
		{name: "email", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.EmailType}, v: "joep@ftv.us", want: "joep@ftv.us"},
		{name: "integer", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.IntegerType}, v: "-987654", want: int64(-987654)},
		{name: "unsigned integer", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.UnsignedIntegerType}, v: "7777777", want: uint64(7777777)},
		{name: "float", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.FloatType}, v: "12.345678", want: 12.345678},
		{name: "boolean", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.BooleanType}, v: "true", want: true},
		{name: "date", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.DateType}, v: "20251012", want: time.Date(2025, 10, 12, 0, 0, 0, 0, time.UTC)},
		{name: "time", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.TimeType}, v: "15:26:37", want: time.Date(0, 1, 1, 15, 26, 37, 0, time.UTC)},
		{name: "timestamp", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.DateTimeType}, v: "2025-10-12 15:26:37.999", want: time.Date(2025, 10, 12, 15, 26, 37, 999000000, time.UTC)},
		{name: "invalid type", f: &Field{Object: Object{Parent: Parent{ID: "f1"}}, Type: 199}, v: "hello world", want: "hello world"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.f.ConvertValue(tc.v)
			assert.Equal(t, tc.want, got)
		})
	}
}
