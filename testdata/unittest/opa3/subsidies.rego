package doelbinding.subsidies

import rego.v1

default allow := false

allow if {
    input.action.attributes.method == "POST"
}

context := {"filter": [{"diplomatiek_kenteken": false}]} if allow

context := {"reason": "alleen POST"} if not allow
