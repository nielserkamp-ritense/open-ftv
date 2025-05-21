package models

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func fieldAsRow(f *schema.Field) *Row {
	def := &schema.Object{Fields: []*schema.Field{}}
	out := &Row{Data: make(map[string]any), def: def}

	out.Data["fqdn"] = f.FQDN()
	def.Fields = append(def.Fields, &fieldDefFQDN)
	out.Data["id"] = f.ID
	def.Fields = append(def.Fields, &fieldDefID)
	out.Data["type"] = f.Type
	def.Fields = append(def.Fields, &fieldDefType)

	if f.Description != "" {
		out.Data["description"] = f.Description
		def.Fields = append(def.Fields, &fieldDefDescription)
	}
	if f.IsArray {
		out.Data["is-array"] = f.IsArray
		def.Fields = append(def.Fields, &fieldDefArray)
	}
	if f.IsEnum {
		out.Data["is-enum"] = f.IsEnum
		def.Fields = append(def.Fields, &fieldDefEnum)
	}
	if f.IsPII {
		out.Data["is-pii"] = f.IsPII
		def.Fields = append(def.Fields, &fieldDefPII)
	}
	if f.Format != "" {
		out.Data["format"] = f.Format
		def.Fields = append(def.Fields, &fieldDefFormat)
	}
	if f.MinLen != 0 {
		out.Data["minimum-length"] = f.MinLen
		def.Fields = append(def.Fields, &fieldDefMinLen)
	}
	if f.MaxLen != 0 {
		out.Data["maximum-length"] = f.MaxLen
		def.Fields = append(def.Fields, &fieldDefMaxLen)
	}
	if f.MinValue != nil {
		out.Data["minimum-value"] = f.MinValue
		def.Fields = append(def.Fields, &fieldDefMinValue)
	}
	if f.MaxValue != nil {
		out.Data["maximum-value"] = f.MaxValue
		def.Fields = append(def.Fields, &fieldDefMaxValue)
	}
	if len(f.AllowedValues) > 0 {
		out.Data["allowed-values"] = f.AllowedValues
		def.Fields = append(def.Fields, &fieldDefAllowed)
	}

	if f.Fields != nil {
		out2 := make([]*Row, 0, len(f.Fields))

		for _, f2 := range f.Fields {
			out2 = append(out2, fieldAsRow(f2))
		}

		out.Data["fields"] = out2
		def.Fields = append(def.Fields, &fieldDefFields)
	}

	return out
}
