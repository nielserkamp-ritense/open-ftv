package doelbinding.zorgtoeslag

import rego.v1

default allow := false

service := data.entities.service[input.uri]

allow if {
    input.http.path == "/organizations"
    input.http.method == "GET"

    service.code == "deelnemers-backend"
    service.owner == "FDS"
}

allow if {
    input.http.path == "/organizations"
    input.http.method == "GET"

    service.code == "catalogus-backend"
    service.owner == "FDS"
}
