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
