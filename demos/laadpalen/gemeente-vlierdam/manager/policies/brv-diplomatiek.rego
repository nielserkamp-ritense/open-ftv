package doelbinding["laadinfra-verificatie"]

import rego.v1

default allow := false

default reason := ""

allow if not input.context.response

allow if input.context.response.status >= 300

allow if {
	is_array(input.context.response.body)
	not has_diplomatic_vehicle
}

allow if {
	input.context.response
	is_gemeentesecretaris
}

has_diplomatic_vehicle if {
	some voertuig in input.context.response.body
	voertuig.diplomatiek == true
}

is_gemeentesecretaris if data.entities.medewerker[input.context.core_user].afdeling == "Gemeentesecretaris"

reason := "geen toestemming om diplomatiek voertuig in te zien, afdeling is geen Gemeentesecretaris" if not allow
