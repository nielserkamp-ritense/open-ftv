# ADR 0002 — Concrete request-mappers als version-tracked plugins

**Status:** geaccepteerd · **Datum:** 2026-07-09

## Context

OpenFTV-core (`eam/mapping`) bevatte naast het generieke `Mapper`-type ook drie
concrete request-mappers — `DoelbindingToPrincipal`, `RvvaToPrincipal` en
`BodyToContext` — plus een hardgecodeerde `switch` (`MappingsFromConfig`) die
config-namen aan die functies koppelde. Dat zijn implementatie-specifieke
voorzieningen: ze coderen keuzes van een specifieke afnemer/inrichting
(doelbinding- en RvVA-conventies, body-parsing) en horen niet in de kern.

ADR 0001 (punt 4) hield de heroverweging van deze mappers expliciet open als een
separaat besluit. Dat besluit wordt hierbij genomen. De actieve configuratie van
mappers wordt sinds de ADL level 4-uitbreiding vastgelegd in
`adl.core.configuration` (`request_mappings`); een geconfigureerde maar
niet-geladen mapper mag daarom nooit stil wegvallen.

## Besluit

1. **Core houdt alleen het mechanisme.** `eam/mapping` behoudt het `Mapper`-type,
   de `Option`/`base`-helpers, de `FromHeaders`-helper en een **registry**:
   - `Register(name string, m Mapper)` — init-time, case-insensitief, dubbele of
     lege naam of nil-mapper = `panic` met duidelijke melding.
   - `Resolve(config string) ([]Mapper, []string)` — resolvet komma-gescheiden
     namen tegen de registry en geeft de ONBEKENDE namen als tweede resultaat
     terug.
   - `MappingsFromConfig(config string) []Mapper` blijft bestaan maar is
     **deprecated** (behoudt de oude signatuur zodat bestaande callers blijven
     compileren; laat onbekende namen stil vallen — gebruik `Resolve`).
2. **De concrete mappers worden version-tracked plugin-modules** onder `plugins/`,
   elk met eigen `go.mod`, tests en README, opgenomen in `go.work`:
   - `plugins/mapping-doelbinding` — `DoelbindingToPrincipal` + `RvvaToPrincipal`
     (dezelfde principal-substitutie-familie; gedeelde `pep`/`convert`-basis;
     kleinste zinvolle versioneringseenheid).
   - `plugins/mapping-body` — `BodyToContext`.
   Elke plugin registreert zich via `init()` onder de bestaande config-namen
   (`doelbindingtoprincipal`, `rvvatoprincipal`, `bodytocontext`).
3. **De gebruikende app importeert plugins expliciet** (blank/underscore-import).
   Die import is een zichtbare, geversioneerde afhankelijkheid in de `go.mod` van
   de app.
4. **`BodyToContext` is gehard.** Een aanwezige maar niet-parsebare body of een
   ontbrekende/onbruikbare `Content-Type` zet niet langer stilzwijgend geen body,
   maar zet een expliciet context-attribuut `body_error`. Body-matchende policies
   moeten default-deny geschreven worden (zie plugin-README).

## Gevolgen

- **Config-namen zijn ongewijzigd.** Bestaande `REQUEST_MAPPINGS`-waarden werken
  ongewijzigd, mits de bijbehorende plugin geïmporteerd is.
- **Niet-geïmporteerde plugin = expliciete warning.** `Resolve` geeft onbekende
  namen terug; de app hoort daarop een waarschuwing/fout te loggen. Een
  geconfigureerde maar niet-geladen mapper valt zo nooit stil weg.
- **Relatie met ADL.** De set actieve mappers blijft verantwoordbaar via
  `adl.core.configuration` (`request_mappings`); de expliciete plugin-import maakt
  de herkomst van elke transformatie traceerbaar en reproduceerbaar.
- Core `eam/mapping` heeft geen afhankelijkheid meer op `eam/pep` of
  `utilities/decode`; die verhuizen mee naar de plugins.
