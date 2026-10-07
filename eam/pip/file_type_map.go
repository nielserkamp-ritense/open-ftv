package pip

func containsAttribute(data map[string]any) bool {
	return data != nil && data["key"] != nil && data["value"] != nil
}

func containsEntity(data map[string]any) bool {
	return data != nil && data["type"] != nil && data["id"] != nil
}

func containsRelation(data map[string]any) bool {
	return data != nil && data["subject"] != nil && data["predicate"] != nil && data["object"] != nil
}

func containsJSONLD(data map[string]any) bool {
	return data != nil && (data["@context"] != nil || data["@id"] != nil || data["@graph"] != nil)
}
