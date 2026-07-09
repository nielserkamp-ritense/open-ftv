# ADR 0001 — Geen server-side normalisatie van AuthZEN-verzoeken

**Status:** geaccepteerd · **Datum:** 2026-07-09

## Context

Voor het FTV-informatiemodel ("proper AuthZEN" conform het NLGov-profiel) horen
`subject.type`, `resource.type` en `action.name` Linked Data URI's te zijn uit de
gepubliceerde FTV-ontologie (`resources/ontology/`). Tijdens de bouw is tijdelijk een
`NormalizeToURI`-request-mapper toegevoegd (via `REQUEST_MAPPINGS`) die binnenkomende
korte namen server-side verrijkte met canonieke URI's.

De werkgroep Federatieve Toegangsverlening heeft eerder als lijn vastgesteld dat
verzoeken en antwoorden in principe niet herschreven mogen worden: de verantwoording
(ADL) moet tonen wat de afnemer daadwerkelijk vroeg en waarop daadwerkelijk is beslist.
Elke server-side transformatie tussen ontvangst en evaluatie ondergraaft die
uitlegbaarheid en maakt replay/audit afhankelijk van transformatie-configuratie.

## Besluit

1. **De afnemer/PEP levert het verzoek canoniek aan** (URI's, `context.mim`,
   `context.ld-context`) conform NLGov AuthZEN. Dat is het stelsel-eindbeeld én de
   productnorm.
2. **OpenFTV normaliseert niet server-side.** De `NormalizeToURI`-mapper is verwijderd
   uit `eam/mapping` en uit de `REQUEST_MAPPINGS`-registratie.
3. De vertaaltabellen kort ↔ URI blijven als bibliotheekfuncties in
   `eam/models/informatiemodel.go` beschikbaar voor PEP's, clients en tooling:
   normalisatie hoort aan de aanleverende kant, vóór verzending.
4. De resterende mappers (`DoelbindingToPrincipal`, `RvvaToPrincipal`, `BodyToContext`)
   blijven vooralsnog bestaan als expliciet te configureren voorzieningen; hun actieve
   configuratie wordt sinds de ADL level 4-uitbreiding vastgelegd in
   `adl.core.configuration` (`request_mappings`), zodat elke resterende transformatie
   verantwoordbaar en reproduceerbaar is. Heroverweging van deze mappers is een
   separaat besluit.

## Gevolgen

- Policies matchen op wat de afnemer stuurt; niet-canonieke verzoeken worden niet
  "gered" door de server. Bestaande demo-policies op korte namen werken alleen met
  niet-canonieke verzoeken (gemarkeerd als legacy in `testdata/authzen/informatiemodel/`).
- De ADL-registratie van verzoek en beslissing is 1-op-1 met wat er binnenkwam.
- Migratiepad voor bestaande afnemers: normaliseren in de eigen PEP/outway met de
  meegeleverde bibliotheekfuncties.
