package models

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"

func fkAsRow(fk *schema.ForeignKey) *Row {
	def := &schema.Object{Fields: []*schema.Field{
		&fieldDefFQDN,
		&fieldDefID,
		&fieldDefDescription,
		&fieldDefForeignTable,
		&fieldDefIndexFields,
	}}

	out := &Row{Data: make(map[string]any), def: def}

	out.Data[fieldDefFQDN.ID] = fk.FQDN()
	out.Data[fieldDefID.ID] = fk.ID

	if fk.Description != "" {
		out.Data[fieldDefDescription.ID] = fk.Description
	}
	if fk.ForeignTable != "" {
		out.Data[fieldDefForeignTable.ID] = fk.ForeignTable
	}
	if len(fk.Fields) > 0 {
		out.Data[fieldDefIndexFields.ID] = fk.Fields
	}

	return out
}
