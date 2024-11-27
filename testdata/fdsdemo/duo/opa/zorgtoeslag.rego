package doelbinding.zorgtoeslag

import rego.v1

default allow := false

service := data.entities.service[input.uri]

allow if {
    startswith(input.http.path, "/personen/")
    input.http.method == "GET"

    service.code == "backend"
    service.owner == "DUO"
}
