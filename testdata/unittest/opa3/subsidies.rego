package doelbinding.subsidies

import rego.v1

default allow := false

allow if {
    input.action.attributes.method == "POST"
}

context := {"organisation_id": "12345", "scope": ["read"]}
