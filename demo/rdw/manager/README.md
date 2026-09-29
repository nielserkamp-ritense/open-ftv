This folder contains the details for the Authorization Manager for RDW:

- `data`        = local attributes/entities for authorizing UX API requests.
- `policies`    = local policies for authorizing UX API requests.
- `tags`        = list of tags for the management interface and bundle management.

## Laadpalen demo, scenario 2

`brv-access` refuses any response that contains a diplomatic vehicle (scenario 1). To share them anyway,
replace its diplomatic rule and `has_diplomatic_vehicle` with `allow if input.context.response` in the RDW
manager, and distribute the `pdp1` bundle. Gemeente Vlierdam's `brv-diplomatiek` policy then shows the
vehicle to a Gemeentesecretaris only.
