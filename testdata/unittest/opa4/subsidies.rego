package doelbinding.subsidies

import rego.v1

default allow := false

allow if input.context.zone == "intern"
