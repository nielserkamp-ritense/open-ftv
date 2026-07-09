# FTV AuthZEN informatiemodel — voorbeeldverzoeken

Deze map bevat voorbeeld-AuthZEN Access Evaluation-verzoeken (`POST /authzen/v1/evaluation`)
die het FTV-informatiemodel volgens het **NLGov-profiel** van de OpenID AuthZEN
Authorization API illustreren ("proper AuthZEN").

Kernidee: `subject.type`, `resource.type` en de `action.name` zijn **Linked Data URI's**
uit een gepubliceerde ontologie in plaats van kale strings. De ontologie en JSON-LD-context:

- `resources/ontology/ftv-informatiemodel.ttl` — de FTV-ontologie: acties als hiërarchie
  van AVG-verwerkingshandelingen (`verwerken` → `raadplegen`, `verstrekken` → `doorzenden`, …),
  plus subject- en resource-typen als klassen, elk met een URI in de namespace
  `https://standaarden.overheid.nl/ftv/def/`.
- `resources/ontology/ftv-context.jsonld` — een JSON-LD 1.1-context die een AuthZEN-verzoek
  (subject/action/resource/context, inclusief `processing_activity_id`) naar deze URI's mapt.

## Canoniek aanleveren is de norm

**De afnemer/PEP levert het verzoek canoniek aan** — met URI's, conform het NLGov-profiel.
OpenFTV normaliseert of herschrijft binnenkomende verzoeken **niet** server-side: een
eerdere `NormalizeToURI`-request-mapper is bewust verwijderd (zie
`docs/adr/0001-geen-server-side-normalisatie.md`). Dit volgt de werkgroep-lijn dat
verzoeken en antwoorden niet herschreven horen te worden (transparantie, uitlegbaarheid,
verantwoording): wat geëvalueerd en gelogd wordt, is wat de afnemer daadwerkelijk vroeg.

De vertaaltabellen kort ↔ URI blijven beschikbaar als bibliotheekfuncties in
`eam/models/informatiemodel.go` (`NormalizeActionURI`, `NormalizeSubjectTypeURI`, …)
voor gebruik in **PEP's/clients en tooling** — de normalisatie hoort dáár, vóór verzending.

## Self-describing maken van een verzoek

Het NLGov-profiel maakt een verzoek zelf-beschrijvend via twee `context`-velden:

- `context.mim` — URL naar het meta-informatiemodel (MIM); hier de gepubliceerde ontologie.
- `context.ld-context` — de Linked Data-context: óf een URL naar de JSON-LD-context,
  óf een inline JSON-LD `@context`-object. Hiermee is het volledige verzoek als RDF-graaf
  te interpreteren, en zijn `action.name` en de property-keys herleidbaar tot URI-concepten.

## Voorbeelden

| Bestand | Toelichting |
|---|---|
| `request-canonical.json` | **De aanbevolen vorm.** Volledig canoniek: alle typen/actie als URI, met `context.mim` en `context.ld-context` als URL. Bevat een verwijzing naar het Register van Verwerkingsactiviteiten via `processing_activity_id`. |
| `request-shortnames.json` | **Niet-canoniek (legacy, afgeraden).** Korte namen (`user`, `raadplegen`, `service`). Werkt alleen tegen policies die zelf op korte namen matchen; OpenFTV vertaalt dit níet naar URI's. |
| `request-inline-ldcontext.json` | Canoniek, met een **inline** `context.ld-context` (`@context`-object) i.p.v. een URL, en een `fsc-peer`-subject met W3C `traceparent`. |
