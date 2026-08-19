package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

var (
	fieldDefFQDN         = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "fqdn"}}}
	fieldDefID           = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "id"}}}
	fieldDefDescription  = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "description"}}}
	fieldDefType         = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "type"}}}
	fieldDefArray        = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "isArray"}}, Type: enums.BooleanType}
	fieldDefEnum         = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "isEnum"}}, Type: enums.BooleanType}
	fieldDefPII          = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "isPii"}}, Type: enums.BooleanType}
	fieldDefFormat       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "format"}}}
	fieldDefMinLen       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "minimumLength"}}, Type: enums.IntegerType}
	fieldDefMaxLen       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "maximumLength"}}, Type: enums.IntegerType}
	fieldDefMinValue     = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "minimumValue"}}}
	fieldDefMaxValue     = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "maximumValue"}}}
	fieldDefAllowed      = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "allowedValues"}}, Type: enums.AnyType, IsArray: true}
	fieldDefFields       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "fields"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefIndexFields  = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "fields"}}, Type: enums.StringType, IsArray: true}
	fieldDefIndexOrders  = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "orders"}}, Type: enums.StringType, IsArray: true}
	fieldDefPK           = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "primaryKey"}}, Type: enums.ObjectType}
	fieldDefIndexes      = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "secondaryIndexes"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefFK           = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "foreignKey"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefForeignTable = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "table"}}}
	fieldDefTables       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "tables"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefSources      = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "sources"}}, Type: enums.ObjectType, IsArray: true}
)

// The row definitions below describe the metadata rows.
//
// They are shared and fixed once, as a schema definition is read-only once it has been fixed,
// and metadata rows are exported concurrently.
var (
	dataspaceRowDef  = newRowDef(&fieldDefFQDN, &fieldDefID, &fieldDefDescription, &fieldDefSources)
	datasourceRowDef = newRowDef(&fieldDefFQDN, &fieldDefID, &fieldDefDescription, &fieldDefTables)
	tableRowDef      = newRowDef(&fieldDefFQDN, &fieldDefID, &fieldDefDescription, &fieldDefFields,
		&fieldDefPK, &fieldDefIndexes, &fieldDefFK)
	indexRowDef = newRowDef(&fieldDefFQDN, &fieldDefID, &fieldDefDescription,
		&fieldDefIndexFields, &fieldDefIndexOrders)
	fkRowDef = newRowDef(&fieldDefFQDN, &fieldDefID, &fieldDefDescription,
		&fieldDefForeignTable, &fieldDefIndexFields)
	fieldRowDef    = newRowDef(fieldRowFields()...)
	fieldPIIRowDef = newRowDef(append(fieldRowFields(), &fieldDefPII)...)
)

func fieldRowFields() []*schema.Field {
	return []*schema.Field{
		&fieldDefFQDN,
		&fieldDefID,
		&fieldDefDescription,
		&fieldDefType,
		&fieldDefArray,
		&fieldDefEnum,
		&fieldDefFormat,
		&fieldDefMinLen,
		&fieldDefMaxLen,
		&fieldDefMinValue,
		&fieldDefMaxValue,
		&fieldDefAllowed,
		&fieldDefFields,
	}
}

func newRowDef(fields ...*schema.Field) *schema.Object {
	def := &schema.Object{Fields: fields}
	def.FixFields(nil, nil)

	return def
}
