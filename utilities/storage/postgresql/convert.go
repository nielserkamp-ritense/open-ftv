package postgresql

import "github.com/google/uuid"

// AnyToUUID converts a PostgreSQL UUID to a readable string.
func AnyToUUID(in any) string {
	switch t := in.(type) {
	case uuid.UUID:
		return t.String()

	case string:
		return t

	case []byte:
		u, _ := uuid.FromBytes(t)
		return u.String()

	case [16]uint8:
		b1, _ := in.([16]uint8)

		var b []byte
		for i := range b1 {
			b = append(b, b1[i])
		}

		u, _ := uuid.FromBytes(b)
		return u.String()

	default:
		u := &uuid.UUID{}
		return u.String()
	}
}
