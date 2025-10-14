package authzen

// EntityFromSearch converts a search entity to a standard entity.
func EntityFromSearch(s *SearchEntity) *Entity {
	return &Entity{Id: s.Id, Properties: s.Properties, Type: s.Type}
}

// EntityToSearch converts a standard entity to a search entity.
func EntityToSearch(s *Entity) *SearchEntity {
	return &SearchEntity{Id: s.Id, Properties: s.Properties, Type: s.Type}
}
