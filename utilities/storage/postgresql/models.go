package postgresql

// Statement represents the details of a single database operation (CREATE, UPDATE or DELETE).
type Statement struct {
	SQL  string
	Args []any
}
