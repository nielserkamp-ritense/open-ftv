# mapping-body

Request-mapper plugin voor OpenFTV. Levert één mapper:

| Config-naam (`REQUEST_MAPPINGS`) | Mapper | Semantiek |
| --- | --- | --- |
| `bodytocontext` | `BodyToContext` | Detecteert een request-body in de action-attributen, decodeert die (base64 → anders letterlijk), parset die met de `Content-Type`-header (of via content-sniffing) en zet het resultaat als context-attribuut `body`, zodat policies op de body-inhoud kunnen matchen. |

## Hardening en default-deny

Deze mapper gaat NIET stilzwijgend door wanneer een aanwezige body niet te
verwerken is. Twee gevallen leiden tot een expliciete fout in plaats van een
ontbrekend body-attribuut:

1. een aanwezige body die niet parsebaar is;
2. een ontbrekende of onbruikbare `Content-Type` waarbij ook content-sniffing
   geen parser oplevert.

In beide gevallen wordt het context-attribuut `body` NIET gezet, maar wordt in
plaats daarvan `body_error` gezet met een reden. Een lege of ontbrekende body
(geen body in het verzoek) is geen fout en laat het verzoek ongewijzigd.

> **Belangrijk:** policies die op de body matchen MOETEN default-deny geschreven
> worden. Een afwezig `body`-attribuut (bijvoorbeeld door `body_error`) mag nooit
> als een impliciete allow worden geïnterpreteerd. Neem `body_error` desgewenst
> expliciet mee als weigergrond.

## Registratie

Importeren van dit pakket registreert de mapper via `init()` bij de core registry
(`eam/mapping`). Een gebruikende app neemt de plugin op met een blank/underscore-import:

```go
import _ "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-body"
```

Zonder deze import is `bodytocontext` niet geregistreerd; `mapping.Resolve` geeft
de naam dan terug als onbekend zodat de app een expliciete waarschuwing kan loggen.

## Versionering

Semver per module, conform de repo-conventie (`semver.sh`, git-tags). De
config-naam `bodytocontext` en het gedrag rond `body_error` zijn onderdeel van het
publieke contract.
