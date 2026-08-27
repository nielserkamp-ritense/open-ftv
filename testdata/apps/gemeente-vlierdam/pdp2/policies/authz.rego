package authz

import rego.v1

default allow := false

default reason := ""

# An aanvraag-beslissing that doesn't involve a diplomatic vehicle needs no special authorization.
allow if {
	input.resource.type == "aanvraag-beslissing"
	input.context.involvesDiplomaticVehicle == false
}

# A diplomatic vehicle may only be decided on by someone from the Gemeentesecretaris department.
allow if {
	input.resource.type == "aanvraag-beslissing"
	input.context.involvesDiplomaticVehicle == true
	input.subject.attributes.afdeling == "Gemeentesecretaris"
}

reason := "diplomatiek kenteken, afdeling is geen Gemeentesecretaris" if {
	not allow
	input.context.involvesDiplomaticVehicle == true
}
