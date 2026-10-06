package pip

// containsAttribute detects if the given data map is likely to contains an attribute.
func containsAttribute(data map[string]any) bool {
	return data != nil && data["key"] != nil && data["value"] != nil
}

// containsEntity detects if the given data map is likely to contains an entity.
func containsEntity(data map[string]any) bool {
	return data != nil && data["type"] != nil && data["id"] != nil
}

// containsRelation detects if the given data map is likely to contains a relation.
func containsRelation(data map[string]any) bool {
	return data != nil && data["subject"] != nil && data["predicate"] != nil && data["object"] != nil
}

// containsJSONLD detects if the given data map is likely to contains JSON-LD formatted RDF data.
func containsJSONLD(data map[string]any) bool {
	return data != nil && (data["@context"] != nil || data["@id"] != nil || data["@graph"] != nil)
}
