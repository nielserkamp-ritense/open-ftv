This folder contains the details for the Policy Decision Point (#2) for gemeente Vlierdam:

- `data`        = local attributes/entities for authorizing UX API requests.
- `policies`    = local policies for authorizing UX API requests.
- `pull-config` = PIP-data pull configurations.

Like pdp1, this PDP's bundle is tagged `rvig`/`rdw` and is managed from the same
`vlierdam-manager`. The only real difference is the policy it carries and the language it's
written in: pdp1 runs Cedar (gates laadpaal-aanvraag submissions, Scenario 1, and authorizes
`vlierdam-outway`'s outbound calls to rvig/rdw); pdp2 runs Rego (gates the diplomat-decision
check, Scenario 2) — demonstrating that one manager can publish policy sets in different
languages to different PDPs.
