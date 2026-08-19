package memory

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
)

// newMockFDS builds a complete mock dataspace.
//
// Every test gets its own copy, as a schema is fixed once, while it is being loaded,
// and is read-only after that.
func newMockFDS() *schema.Dataspace {
	var (
		bsn        = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "bsn"}, Description: "Burgerservicenummer"}, IsPII: true, Type: enums.StringType}
		voornaam   = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "voornaam"}, Description: "Voornaam"}, IsPII: true, Type: enums.StringType}
		achternaam = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "achternaam"}, Description: "Achternaam"}, IsPII: true, Type: enums.StringType}
		postcode   = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "postcode"}, Description: "Postcode"}, Type: enums.StringType}
		kenteken   = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "kenteken"}, Description: "kenteken"}, IsPII: true, Type: enums.StringType}

		f1 = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.IntegerType}
		f2 = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.StringType}
		f3 = &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.BooleanType}

		t1 = &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "persoon"},
				Description: "Personen",
				Fields:      []*schema.Field{bsn, voornaam, achternaam},
			},
			PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"bsn"}},
			SecondaryIndexes: []*schema.Index{
				{Parent: schema.Parent{ID: "ix2"}, Description: "Voornamen", Fields: []string{"voornaam"}},
				{Parent: schema.Parent{ID: "ix3"}, Description: "Achternamen", Fields: []string{"achternaam"}},
			},
		}

		t2 = &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "adres"},
				Description: "Adressen",
				Fields:      []*schema.Field{bsn, postcode},
			},
			PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"bsn"}},
			ForeignKeys: []*schema.ForeignKey{
				{
					Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
					ForeignTable: "persoon",
				},
			},
		}

		t3 = &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "kenteken"},
				Description: "Kentekens",
				Fields:      []*schema.Field{kenteken, bsn},
			},
			PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"kenteken"}},
		}

		t4 = &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "adres"},
				Description: "Adressen",
				Fields:      []*schema.Field{bsn, postcode},
			},
			PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"bsn"}},
			ForeignKeys: []*schema.ForeignKey{
				{
					Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
					ForeignTable: "kenteken",
				},
			},
		}

		dummy = &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "dummy"},
				Description: "Dummy table without data",
				Fields:      []*schema.Field{f1, f2, f3},
			},
			PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"f1"}},
		}

		ds1 = &schema.Datasource{
			Parent:      schema.Parent{ID: "brp"},
			Description: "Basis Registratie Personen",
			Tables:      []*schema.Table{t1, t2, dummy},
		}

		ds2 = &schema.Datasource{
			Parent:      schema.Parent{ID: "rdw"},
			Description: "Rijksdienst voor het Wegverkeer",
			Tables:      []*schema.Table{t3, t4},
		}

		mockFDS = &schema.Dataspace{
			Parent:      schema.Parent{ID: "fds"},
			Description: "Federatief DataStelsel",
			DataSources: []*schema.Datasource{ds1, ds2},
		}
	)

	return mockFDS
}

// newMockStore builds a loaded store over a fresh mock dataspace,
// and returns it together with the "brp" datasource definition most tests filter on.
func newMockStore(t *testing.T) (store.Storage, *schema.Datasource) {
	t.Helper()

	fds := newMockFDS()
	s := New(fds)
	require.NotNil(t, s)

	loadFDS(s)

	return s, fds.Source("brp")
}
