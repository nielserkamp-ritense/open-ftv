package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/writer/csv"
)

func writeKeys(rec *Row, enc csv.BytesEncoder) {
	rec.def.IterateFields(func(def *schema.Field) {
		enc.WriteString(def.ID)
		enc.WriteSeparator()
	})
	enc.WriteEOL()
}

func writeRecord(rec *Row, enc csv.BytesEncoder) {
	rec.def.IterateFields(func(def *schema.Field) {
		v := rec.Data[def.ID]
		enc.Encode(def, v)
		enc.WriteSeparator()
	})
	enc.WriteEOL()
}
