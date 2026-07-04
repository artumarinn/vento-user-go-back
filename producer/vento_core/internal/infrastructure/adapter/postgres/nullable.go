package postgres

// nullableUUID converts an empty string into a nil driver value so it binds
// to a nullable UUID column as SQL NULL instead of failing UUID parsing.
func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
