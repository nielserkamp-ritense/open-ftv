package decisions

// Status describes the outcome of a PDP evaluation attempt (Logius ADL).
type Status string

const (
	StatusUnset Status = "Unset"
	StatusOk    Status = "Ok"
	StatusError Status = "Error"
)
