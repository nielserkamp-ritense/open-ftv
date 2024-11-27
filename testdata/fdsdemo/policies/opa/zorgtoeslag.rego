package doelbinding.zorgtoeslag

import rego.v1

default allow := false

service := data.entities.service[input.uri]

allow if {
    input.http.path == "/networth"
    input.http.method == "GET"

    service.code == "backend"
    service.owner == "Belastingdienst"
}
