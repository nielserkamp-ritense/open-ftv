package Doelbinding.subsidies

import rego.v1

default allow := false

allow if {
    input.http.method == "POST"
}
