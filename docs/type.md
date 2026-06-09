# PEP, PDP, PIP & PAP

Standaard access-control rollen (XACML / AuthZEN), zoals OpenFTV ze gebruikt.

## Overzicht

| Afkorting | Voluit | Wat het doet | In OpenFTV |
|---|---|---|---|
| **PEP** | Policy **Enforcement** Point | Onderschept de request, bouwt de autorisatievraag (PARC: principal / action / resource / context), vraagt de PDP om een oordeel en **handhaaft** dat (doorlaten of `403`). | `eam/pep` — valideert het JWT, mapt token → principal + rollen, bouwt de PARC. Draait in de gateway-plugin (Kong) én in-app in de `manager`. |
| **PDP** | Policy **Decision** Point | Neemt de **beslissing** (`permit` / `deny`) door de PARC tegen de policies te evalueren. Handhaaft zelf niets. | `eam/pdp/cedar-embedded` — de embedded Cedar-engine die `permit` / `deny` teruggeeft. |
| **PAP** | Policy **Administration** Point | **Beheert** de policies: aanmaken, bewerken, opslaan, versioneren, uitleveren. De bron van de regels. | `eam/pap` — de policy-store (file-store vs. postgres). De UI ("Beleidsregels") praat hiermee. |
| **PIP** | Policy **Information** Point | Levert **extra attributen / context** die de PDP nodig heeft maar niet in de request zitten (rollen, resource-status, eigenaarschap, …). | `eam/pip` — de entity-/attribuut-store (`PIP_STORE`). |

## Flow

1. Request komt binnen → de **PEP** vangt 'm en bepaalt principal / action / resource.
2. De **PIP** vult ontbrekende attributen aan (rollen, resource-status, …).
3. De **PDP** evalueert de PARC tegen de policies → `permit` / `deny`.
4. De **PEP** handhaaft de uitkomst (doorlaten of `403`).
5. De **PAP** is waar die policies vandaan komen en beheerd worden.

## Ezelsbruggetje

> **PEP** *enforces* · **PDP** *decides* · **PIP** *informs* · **PAP** *administers*

## Relevante nuance (OpenFTV)

- **Fijnmazige policies** (bijv. `resource.status == "draft"`) werken doordat de PEP de
  opgeslagen resource — met zijn attributen — meegeeft in de PARC, zodat de PDP daar
  Cedar-regels tegenaan kan evalueren.
- **Rollen** komen uit de `roles`-claim van het access-token (configureerbaar via
  `OIDC_ROLES_CLAIM`) en landen als principal-attribuut in de PARC.
- De **PAP is de bron van waarheid**: bij een postgres-backed store worden de gebundelde
  Cedar-policies éénmalig geseed (met UUID-ids); daarna winnen UI-wijzigingen.
