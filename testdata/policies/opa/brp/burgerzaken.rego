package Doelbinding.burgerzaken

import rego.v1

default allow := false

principal := data.entities.Doelbinding.burgerzaken.brpPersonen

allow if {
    is_brp
    input.http.method == "POST"
    input.body.type in principal.types

    fields := input.body.fields
    every _, f in fields {
        f in allowed_fields
    }
}

is_brp if {
    service := data.entities.Service[input.uri]
    service.code == "BRP"
    service.owner == "RvIG"
}

allowed_fields := principal.zoeken.allowedFields if {
    startswith(input.body.type, "Zoek")
}

allowed_fields := principal.raadplegen.allowedFields if {
    startswith(input.body.type, "Raadpleeg")
}
