package handlers

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

func TestEntityFromOAS(t *testing.T) {
	testCases := []struct {
		name string
		in   *attributes.Entity
		want models.Entity
	}{
		{
			name: "no attributes",
			in:   &attributes.Entity{Type: "type", Id: "id"},
			want: models.NewEntity("type", "id", models.NewAttributeSet()),
		},
		{
			name: "one attribute",
			in: &attributes.Entity{
				Type: "type",
				Id:   "id",
				Attributes: []attributes.Attribute{
					{Key: "key1", Value: "value", Type: "xsd:string"},
				},
			},
			want: models.NewEntity(
				"type",
				"id",
				models.NewAttributeSet(
					models.NewAttribute("key1", "value"),
				),
			),
		},
		{
			name: "few attributes",
			in: &attributes.Entity{
				Type: "type",
				Id:   "id",
				Attributes: []attributes.Attribute{
					{Key: "key1", Value: "value", Type: "xsd:string"},
					{Key: "key2", Value: "123.456", Type: "xsd:float"},
					{Key: "key3", Value: "123456", Type: "xsd:nonNegativeInteger"},
				},
			},
			want: models.NewEntity(
				"type",
				"id",
				models.NewAttributeSet(
					models.NewAttribute("key1", "value"),
					models.NewAttribute("key2", 123.456),
					models.NewAttribute("key3", int64(123456)),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := EntityFromOAS(tc.in, models.NewAttributeSet())
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestEntityToOAS(t *testing.T) {
	testCases := []struct {
		name string
		in   models.Entity
		want *attributes.Entity
	}{
		{
			name: "no attributes",
			in:   models.NewEntity("type", "id", models.NewAttributeSet()),
			want: &attributes.Entity{Type: "type", Id: "id", Attributes: []attributes.Attribute{}},
		},
		{
			name: "one attribute",
			in: models.NewEntity(
				"type",
				"id",
				models.NewAttributeSet(
					models.NewAttribute("key1", "value"),
				),
			),
			want: &attributes.Entity{
				Type: "type",
				Id:   "id",
				Attributes: []attributes.Attribute{
					{Key: "key1", Value: "value", Type: "xsd:string"},
				},
			},
		},
		{
			name: "few attributes",
			in: models.NewEntity(
				"type",
				"id",
				models.NewAttributeSet(
					models.NewAttribute("key3", int64(123456)),
					models.NewAttribute("key1", "value"),
					models.NewAttribute("key2", 123.456),
				),
			),
			want: &attributes.Entity{
				Type: "type",
				Id:   "id",
				Attributes: []attributes.Attribute{
					{Key: "key1", Value: "value", Type: "xsd:string"},
					{Key: "key2", Value: 123.456, Type: "xsd:double"},
					{Key: "key3", Value: int64(123456), Type: "xsd:long"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := EntityToOAS(tc.in)

			slices.SortFunc(got.Attributes, func(a, b attributes.Attribute) int {
				return cmp.Compare(a.Key, b.Key)
			})

			assert.EqualValues(t, tc.want, got)
		})
	}
}
