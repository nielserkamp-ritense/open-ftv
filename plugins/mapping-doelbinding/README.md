# mapping-doelbinding

Request-mapper plugin voor OpenFTV. Levert twee mappers uit de
"principal-substitutie op grond van een grondslag-attribuut"-familie:

| Config-naam (`REQUEST_MAPPINGS`) | Mapper | Semantiek |
| --- | --- | --- |
| `doelbindingtoprincipal` | `DoelbindingToPrincipal` | Als er een *doelbinding* aanwezig is (context-attribuut `doelbinding` of header `doelbinding`) en de principal nog geen doelbinding is, wordt de doelbinding de principal. De oorspronkelijke principal blijft bewaard in `context.client_principal`. |
| `rvvatoprincipal` | `RvvaToPrincipal` | Idem, maar voor een RvVA-ID (context-attribuut `rvva_id` of header `dpl-processing-activity-id` / `x-dpl-rva-activity-id`). |

Beide mappers zitten in één module omdat ze functioneel samenhangen (identieke
structuur, dezelfde `pep`-/`convert`-afhankelijkheden en dezelfde semantiek).
De repo versioneert per module, niet per functie; deze functioneel-samenhangende
eenheid is de kleinste zinvolle versioneringseenheid.

## Registratie

Importeren van dit pakket registreert beide mappers via `init()` bij de core
registry (`eam/mapping`). Een gebruikende app neemt de plugin op met een
blank/underscore-import:

```go
import _ "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-doelbinding"
```

Zonder deze import zijn `doelbindingtoprincipal` en `rvvatoprincipal` niet
geregistreerd; `mapping.Resolve` geeft ze dan terug als onbekende namen zodat de
app een expliciete waarschuwing kan loggen.

## Versionering

Semver per module, conform de repo-conventie (`semver.sh`, git-tags). De
config-namen (`doelbindingtoprincipal`, `rvvatoprincipal`) zijn onderdeel van het
publieke contract: een wijziging daarvan is een major bump.
