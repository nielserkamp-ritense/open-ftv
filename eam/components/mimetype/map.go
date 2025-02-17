package mimetype

// ContainsAttribute detects if the given data map is likely to contains an attribute.
func ContainsAttribute(data map[string]any) bool {
	return data != nil && data["key"] != nil && data["value"] != nil
}

// ContainsEntity detects if the given data map is likely to contains an entity.
func ContainsEntity(data map[string]any) bool {
	return data != nil && data["type"] != nil && data["id"] != nil
}

// ContainsRelation detects if the given data map is likely to contains a relation.
func ContainsRelation(data map[string]any) bool {
	return data != nil && data["subject"] != nil && data["predicate"] != nil && data["object"] != nil
}

// ContainsJSONLD detects if the given data map is likely to contains JSON-LD formatted RDF data.
func ContainsJSONLD(data map[string]any) bool {
	return data != nil && (data["@context"] != nil || data["@id"] != nil || data["@graph"] != nil)
}
