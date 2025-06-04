package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

var (
	fieldDefFQDN         = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "fqdn"}}}
	fieldDefID           = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "id"}}}
	fieldDefDescription  = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "description"}}}
	fieldDefType         = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "type"}}}
	fieldDefArray        = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "is-array"}}, Type: enums.BooleanType}
	fieldDefEnum         = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "is-enum"}}, Type: enums.BooleanType}
	fieldDefPII          = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "is-pii"}}, Type: enums.BooleanType}
	fieldDefFormat       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "format"}}}
	fieldDefMinLen       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "minimum-length"}}, Type: enums.IntegerType}
	fieldDefMaxLen       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "maximum-length"}}, Type: enums.IntegerType}
	fieldDefMinValue     = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "minimum-value"}}}
	fieldDefMaxValue     = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "maximum-value"}}}
	fieldDefAllowed      = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "allowed-values"}}, Type: enums.AnyType, IsArray: true}
	fieldDefFields       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "fields"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefIndexFields  = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "fields"}}, Type: enums.StringType, IsArray: true}
	fieldDefIndexOrders  = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "orders"}}, Type: enums.StringType, IsArray: true}
	fieldDefPK           = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "primary-key"}}, Type: enums.ObjectType}
	fieldDefIndexes      = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "secondary-indexes"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefFK           = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "foreign-key"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefForeignTable = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "table"}}}
	fieldDefTables       = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "tables"}}, Type: enums.ObjectType, IsArray: true}
	fieldDefSources      = schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "sources"}}, Type: enums.ObjectType, IsArray: true}
)
