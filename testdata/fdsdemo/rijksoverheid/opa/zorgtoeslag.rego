package doelbinding.zorgtoeslag

import rego.v1

default allow := false

service := data.entities.service[input.uri]

allow if {
    input.http.path == "/haalcentraal/api/brp/personen"
    input.http.method == "POST"
    input.body.type == "RaadpleegMetBurgerservicenummer"

    service.code == "brp-backend"
    service.owner == "Rijksoverheid"
}

allow if {
    startswith(input.http.path, "/personen/")
    input.http.method == "POST"

    service.code == "berichtenbox-backend"
    service.owner == "Rijksoverheid"
}
