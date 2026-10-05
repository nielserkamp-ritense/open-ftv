package doelbinding["identificatie-en-ingezetenschap"]

import rego.v1

default allow := false

default reason := ""

is_brp_request if input.context.http.path == "/v1/haalcentraal/api/brp/personen"

allow if not is_brp_request

has_correct_doelbinding if input.context.doelbinding == "identificatie-en-ingezetenschap"

allow if {
	is_brp_request
	has_correct_doelbinding
}

reason := "verkeerde doelbinding" if {
	is_brp_request
	not has_correct_doelbinding
}
