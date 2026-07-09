# ODRL-Geo-NL engine test data

Real ODRL-Geo-NL example policies from the BRO simulation profile, used by the
integration tests of this engine:

- `example1.ttl` — grondwatermonitoring, `geonl:featureProperty` sensitivity
  constraint (waivable for government assignees via an Agreement).
- `example2.ttl` — Defensie area prohibition (`geof:sfWithin` on a WKT polygon
  in RD New) plus the equivalent `sfDisjoint` permission and a
  `geonl:LayerSelection` variant.
- `example3.ttl` — "ja, mits" generalisation + non-redistribution duties mapped
  to AuthZEN obligations.
- `layerselection.ttl` — an isolated `geonl:LayerSelection` policy.

## PIP data model for `geonl:LayerSelection`

A `geonl:LayerSelection` (spec.md §3.2, §7.4) selects zone geometries from
another layer via the PIP. The engine (`pip.go`, `resolveLayerSelection`)
expects, for each zone of the referenced layer, a **PIP entity** whose:

- `Type()` equals the `geonl:fromLayer` URI
  (e.g. `https://api.example.gov.nl/geo/milieuzonering`);
- attributes include a `geometry` attribute holding a WKT literal (optionally
  CRS-prefixed), plus every attribute referenced by the selection's
  `geonl:where` filter (e.g. `classificatie`).

The `geonl:where` constraint is applied to each entity's attributes; only
matching zones contribute to the (universally quantified) spatial relation. An
**unresolvable** layer (no entities of that type) makes the spatial constraint
*not evaluable* — fail-safe per spec.md §7.5 — rather than vacuously true.

Example entity (as added in the tests):

    type: https://api.example.gov.nl/geo/milieuzonering
    id:   zone-1
    attributes:
      classificatie: gevoelig
      geometry: "<...EPSG/0/28992> POLYGON ((150000 460000, 158000 460000, 158000 466000, 150000 466000, 150000 460000))"
