package doelbinding.burgerzaken

import rego.v1

default allow = false

allow if input.http.method == "POST"
