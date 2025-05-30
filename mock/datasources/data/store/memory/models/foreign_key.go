package models

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"

func fkAsRow(fk *schema.ForeignKey) *Row {
	def := &schema.Object{Fields: make([]*schema.Field, 0)}
	out := &Row{Data: make(map[string]any), def: def}

	out.Data["fqdn"] = fk.FQDN()
	def.Fields = append(def.Fields, &fieldDefFQDN)
	out.Data["id"] = fk.ID
	def.Fields = append(def.Fields, &fieldDefID)

	if fk.Description != "" {
		out.Data["description"] = fk.Description
		def.Fields = append(def.Fields, &fieldDefDescription)
	}

	if fk.ForeignTable != "" {
		out.Data["table"] = fk.ForeignTable
		def.Fields = append(def.Fields, &fieldDefForeignTable)
	}

	if len(fk.Fields) > 0 {
		out.Data["fields"] = fk.Fields
		def.Fields = append(def.Fields, &fieldDefIndexFields)
	}

	return out
}
