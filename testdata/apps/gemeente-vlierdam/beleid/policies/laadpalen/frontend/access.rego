package authz

default allow := false

default reason := ""

reason := "not a valid subject type" if {
	input.subject.type != "user"
} else := "not a valid user" if {
	input.subject.type == "user"
	not input.subject.id in ["Morty", "Beth", "Jerry"]
} else := "not certified" if {
	input.subject.type == "user"
	input.subject.id == "Morty"
} else := "certification expired" if {
	input.subject.type == "user"
	input.subject.id == "Jerry"
}

allow if {
	input.subject.type == "user"
	input.subject.id == "Beth"
}
