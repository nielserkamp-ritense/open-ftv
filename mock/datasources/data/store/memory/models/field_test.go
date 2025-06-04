package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
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
				Type:   enums.StringType,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f1",
				"id":          "f1",
				"description": "field 1",
				"type":        enums.StringType,
			}},
		},
		{
			name: "array",
			field: &schema.Field{
				Object:  schema.Object{Parent: schema.Parent{ID: "f2"}, Description: "field 2"},
				Type:    enums.IntegerType,
				IsArray: true,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f2",
				"id":          "f2",
				"description": "field 2",
				"type":        enums.IntegerType,
				"is-array":    true,
			}},
		},
		{
			name: "enum",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f3"}, Description: "field 3"},
				Type:   enums.IntegerType,
				IsEnum: true,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f3",
				"id":          "f3",
				"description": "field 3",
				"type":        enums.IntegerType,
				"is-enum":     true,
			}},
		},
		{
			name: "pii",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f4"}, Description: "field 4"},
				Type:   enums.StringType,
				IsPII:  true,
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f4",
				"id":          "f4",
				"description": "field 4",
				"type":        enums.StringType,
				"is-pii":      true,
			}},
		},
		{
			name: "format",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f5"}, Description: "field 5"},
				Type:   enums.DateTimeType,
				Format: "20060102150405",
			},
			want: &Row{Data: map[string]any{
				"fqdn":        "f5",
				"id":          "f5",
				"description": "field 5",
				"type":        enums.DateTimeType,
				"format":      "20060102150405",
			}},
		},
		{
			name: "min/max length",
			field: &schema.Field{
				Object: schema.Object{Parent: schema.Parent{ID: "f6"}, Description: "field 6"},
				Type:   enums.StringType,
				MinLen: 1,
				MaxLen: 20,
			},
			want: &Row{Data: map[string]any{
				"fqdn":           "f6",
				"id":             "f6",
				"description":    "field 6",
				"type":           enums.StringType,
				"minimum-length": 1,
				"maximum-length": 20,
			}},
		},
		{
			name: "min/max value",
			field: &schema.Field{
				Object:   schema.Object{Parent: schema.Parent{ID: "f7"}, Description: "field 7"},
				Type:     enums.UnsignedIntegerType,
				MinValue: 1,
				MaxValue: 10,
			},
			want: &Row{Data: map[string]any{
				"fqdn":          "f7",
				"id":            "f7",
				"description":   "field 7",
				"type":          enums.UnsignedIntegerType,
				"minimum-value": 1,
				"maximum-value": 10,
			}},
		},
		{
			name: "allowed values",
			field: &schema.Field{
				Object:        schema.Object{Parent: schema.Parent{ID: "f8"}, Description: "field 8"},
				Type:          enums.StringType,
				IsEnum:        true,
				AllowedValues: []any{"M", "V", "O"},
			},
			want: &Row{Data: map[string]any{
				"fqdn":           "f8",
				"id":             "f8",
				"description":    "field 8",
				"type":           enums.StringType,
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
						Type:   enums.StringType,
					},
					{
						Object: schema.Object{Parent: schema.Parent{ID: "f2"}, Description: "field 2"},
						Type:   enums.DateType,
					},
					{
						Object: schema.Object{Parent: schema.Parent{ID: "f3"}, Description: "field 3"},
						Type:   enums.BooleanType,
					},
				},
			},
			Type: enums.ObjectType,
		}

		got := fieldAsRow(field)
		require.NotNil(t, got)

		assert.Equal(t, "f9", got.Data["fqdn"])
		assert.Equal(t, "f9", got.Data["id"])
		assert.Equal(t, "field 9", got.Data["description"])
		assert.EqualValues(t, enums.ObjectType, got.Data["type"])
		assert.Len(t, got.Data["fields"], 3)
	})
}
