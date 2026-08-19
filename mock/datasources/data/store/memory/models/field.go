package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

func fieldAsRow(f *schema.Field) *Row {
	out := &Row{Data: make(map[string]any), def: fieldRowDef}

	out.Data[fieldDefFQDN.ID] = f.FQDN()
	out.Data[fieldDefID.ID] = f.ID
	out.Data[fieldDefType.ID] = f.Type

	if f.Description != "" {
		out.Data[fieldDefDescription.ID] = f.Description
	}
	if f.IsArray {
		out.Data[fieldDefArray.ID] = f.IsArray
	}
	if f.IsEnum {
		out.Data[fieldDefEnum.ID] = f.IsEnum
	}
	if f.IsPII {
		out.Data[fieldDefPII.ID] = f.IsPII
		out.def = fieldPIIRowDef
	}
	if f.Format != "" {
		out.Data[fieldDefFormat.ID] = f.Format
	}
	if f.MinLen != 0 {
		out.Data[fieldDefMinLen.ID] = f.MinLen
	}
	if f.MaxLen != 0 {
		out.Data[fieldDefMaxLen.ID] = f.MaxLen
	}
	if f.MinValue != nil {
		out.Data[fieldDefMinValue.ID] = f.MinValue
	}
	if f.MaxValue != nil {
		out.Data[fieldDefMaxValue.ID] = f.MaxValue
	}
	if len(f.AllowedValues) > 0 {
		out.Data[fieldDefAllowed.ID] = f.AllowedValues
	}

	if f.Fields != nil {
		out2 := make([]*Row, 0, len(f.Fields))
		for _, f2 := range f.Fields {
			out2 = append(out2, fieldAsRow(f2))
		}
		out.Data[fieldDefFields.ID] = out2
	}

	return out
}
