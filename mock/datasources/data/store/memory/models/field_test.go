package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestFieldAsRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		field *schema.Field
		want  *Row
	}{
		{
			name: "simple",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f1"}, Description: "field 1"},
				Type:   types.StringType,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f1",
				"id":          "f1",
				"description": "field 1",
				"type":        types.StringType,
			}},
		},
		{
			name: "array",
			field: &schema.Field{
				Object:  schema.Object{Parent: schema.Parent{ID: "f2"}, Description: "field 2"},
				Type:    types.IntegerType,
				IsArray: true,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f2",
				"id":          "f2",
				"description": "field 2",
				"type":        types.IntegerType,
				"is-array":    true,
			}},
		},
		{
			name: "enum",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f3"}, Description: "field 3"},
				Type:   types.IntegerType,
				IsEnum: true,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f3",
				"id":          "f3",
				"description": "field 3",
				"type":        types.IntegerType,
				"is-enum":     true,
			}},
		},
		{
			name: "pii",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f4"}, Description: "field 4"},
				Type:   types.StringType,
				IsPII:  true,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f4",
				"id":          "f4",
				"description": "field 4",
				"type":        types.StringType,
				"is-pii":      true,
			}},
		},
		{
			name: "format",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f5"}, Description: "field 5"},
				Type:   types.DateTimeType,
				Format: "20060102150405",
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f5",
				"id":          "f5",
				"description": "field 5",
				"type":        types.DateTimeType,
				"format":      "20060102150405",
			}},
		},
		{
			name: "min/max length",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f6"}, Description: "field 6"},
				Type:   types.StringType,
				MinLen: 1,
				MaxLen: 20,
			},
			want: &Row{Data: map[string]any{
				"fqdn":           "f6",
				"id":             "f6",
				"description":    "field 6",
				"type":           types.StringType,
				"minimum-length": 1,
				"maximum-length": 20,
			}},
		},
		{
			name: "min/max value",
			field: &schema.Field{
				Object:   schema.Object{Parent: schema.Parent{ID: "f7"}, Description: "field 7"},
				Type:     types.UnsignedIntegerType,
				MinValue: 1,
				MaxValue: 10,
			},
			want: &Row{Data: map[string]any{
				"fqdn":          "f7",
				"id":            "f7",
				"description":   "field 7",
				"type":          types.UnsignedIntegerType,
				"minimum-value": 1,
				"maximum-value": 10,
			}},
		},
		{
			name: "allowed values",
			field: &schema.Field{
				Object:        schema.Object{Parent: schema.Parent{ID: "f8"}, Description: "field 8"},
				Type:          types.StringType,
				IsEnum:        true,
				AllowedValues: []any{"M", "V", "O"},
			},
			want: &Row{Data: map[string]any{
				"fqdn":           "f8",
				"id":             "f8",
				"description":    "field 8",
				"type":           types.StringType,
				"is-enum":        true,
				"allowed-values": []any{"M", "V", "O"},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := fieldAsRow(tc.field)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestObjectFieldAsRecord(t *testing.T) {
	t.Parallel()

	t.Run("object field", func(t *testing.T) {
		t.Parallel()

		field := &schema.Field{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "f9"},
				Description: "field 9",
				Fields: []*schema.Field{
					{
						Object: schema.Object{Parent: schema.Parent{ID: "f1"}, Description: "field 1"},
						Type:   types.StringType,
					},
					{
						Object: schema.Object{Parent: schema.Parent{ID: "f2"}, Description: "field 2"},
						Type:   types.DateType,
					},
					{
						Object: schema.Object{Parent: schema.Parent{ID: "f3"}, Description: "field 3"},
						Type:   types.BooleanType,
					},
				},
			},
			Type: types.ObjectType,
		}

		got := fieldAsRow(field)
		require.NotNil(t, got)

		assert.Equal(t, "f9", got.Data["fqdn"])
		assert.Equal(t, "f9", got.Data["id"])
		assert.Equal(t, "field 9", got.Data["description"])
		assert.EqualValues(t, types.ObjectType, got.Data["type"])
		assert.Len(t, got.Data["fields"], 3)
	})
}
