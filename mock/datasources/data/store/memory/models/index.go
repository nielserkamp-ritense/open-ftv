package models

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"

func indexAsRow(ix *schema.Index) *Row {
	def := &schema.Object{Fields: make([]*schema.Field, 0)}
	out := &Row{Data: make(map[string]any), def: def}

	out.Data["fqdn"] = ix.FQDN()
	def.Fields = append(def.Fields, &fieldDefFQDN)
	out.Data["id"] = ix.ID
	def.Fields = append(def.Fields, &fieldDefID)

	if ix.Description != "" {
		out.Data["description"] = ix.Description
		def.Fields = append(def.Fields, &fieldDefDescription)
	}

	if len(ix.Fields) > 0 {
		out.Data["fields"] = ix.Fields
		def.Fields = append(def.Fields, &fieldDefIndexFields)
	}

	if len(ix.Orders) > 0 {
		out.Data["orders"] = ix.Orders
		def.Fields = append(def.Fields, &fieldDefIndexOrders)
	}

	return out
}
