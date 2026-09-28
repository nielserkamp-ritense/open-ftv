package doelbinding["laadinfra-verificatie"]

import rego.v1

default allow := false

default reason := ""

allow if not input.context.response

allow if input.context.response.status >= 300

# Scenario 1: RDW doesn't share diplomatic vehicles. Scenario 2 replaces this rule and the next with
# `allow if input.context.response`, leaving it to the gemeente to show them to the Gemeentesecretaris only.
allow if {
	is_array(input.context.response.body)
	not has_diplomatic_vehicle
}

has_diplomatic_vehicle if {
	some voertuig in input.context.response.body
	voertuig.diplomatiek == true
}

reason := "RDW verstrekt geen gegevens over diplomatieke voertuigen" if not allow
