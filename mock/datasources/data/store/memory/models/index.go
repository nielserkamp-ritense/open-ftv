package models

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"

func indexAsRow(ix *schema.Index) *Row {
	def := &schema.Object{Fields: []*schema.Field{
		&fieldDefFQDN,
		&fieldDefID,
		&fieldDefDescription,
		&fieldDefIndexFields,
		&fieldDefIndexOrders,
	}}

	out := &Row{Data: make(map[string]any), def: def}

	out.Data[fieldDefFQDN.ID] = ix.FQDN()
	out.Data[fieldDefID.ID] = ix.ID

	if ix.Description != "" {
		out.Data[fieldDefDescription.ID] = ix.Description
	}
	if len(ix.Fields) > 0 {
		out.Data[fieldDefIndexFields.ID] = ix.Fields
	}
	if len(ix.Orders) > 0 {
		out.Data[fieldDefIndexOrders.ID] = ix.Orders
	}

	return out
}
