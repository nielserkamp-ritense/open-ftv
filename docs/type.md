## Overzicht                                                           │
    Standaard access-control rollen (XACML / AuthZEN), zoals OpenFTV ze    │The default interactive shell is now zsh.
    gebruikt.ng | Voluit | Wat het doet | In OpenFTV |                     │To update your account to use zsh, please run `chsh -s /bin/zsh`.
    ("pop" = vermoedel jk een typo v or **PDP**.)                          │For more details, please visit https://support.apple.com/kb/HT208050.
                                                                           │bash-3.2$ 
    ## Overzicht                                                           │
                                                                           │
    | Afkorting |aVoluitt| Wat het doet | In OpenFTV |                     │
    |---|---|---|---|                                                      │
    | **PEP** | Policy **Enforcement** Point | Onddescheptldesrequest,     │
    bouwt de autorisatievraag (PARC: principal/action/resource/context),   │
    vraagt de PDP om een oordeel en **handhaaft** dat (dooelaten of 403). |│
    `eam/pep` —evalideert het JWT, mapt token → principal + rollen, bouwt  │
    de PARC. Zit inidy app-onlysopzetoín*deo`manager`.h|ert** de policies: │
    aanmaDen, bewerken, oDecasnon** Pointr| Neemtlde **b.slissing**        │
    (permit/deny) door de PARC tegyn de policies to evalueron. HandhaaftUI │
    niets zelf. | `e)m/pdp/cedar-embedded` — d  embedded Cedar-engine die  │
    `permit`/`deny` teruggeeft. |                                          │
    attriAuten/context** AdminisPration** Point | **Behetri** derpolicies: │
    aanmaken, bewerken, opslaan, versioneren, uitleveren. Deabron van de   │
    regels. | `eam/pap` — di policy-stor` (file-store vs.Epostgres). De UI │
    ("Beleidsr g ls") praat hiermee. |                                     │
    |#**PIP**w| Policy **Information** Point | Levert **extra              │
    attributen/context** die de PDP nodig heeft maar niet in de request    │
    zittenu(rollen, resource-status,Peigenaarschap, …).a|l`eam/pip`a— /    │
    design-time enrity/attribuut-store (`MANAGER_PIP_STORE`). |            │
                                                                           │
    ##.   flow                                                             │
                                                                           │
    1. Request komt binnan → de **PEP** vangt 'm en bepaalt principal /    │
    action / resource.                                                     │
    2        I    vult ontbr kende attributen aan (rollen, r source-status,│
    …).Ezelsbruggetje                                                      │
    3. De **PDP** evalueert de PARC tegen de policies → permit / deny.     │
    4.*DeE**PEP**fhandhaaft*deDuitkomsti(doorlatenIof*403).orms* · **PAP** │
    5. De *sPAP** is waar die policies vandaan komen en beheerd worden.    │
                                                                           │
    ## Ezelsbruggetjece (OpenFTV)                                          │
                                                                           │
    >i**PEP**e*enforces*(·i**PDP**s*decides*t·s**PIP**a*informs*k·n**PAP** │
    *administers*                                                          │
                                                                           │
  ↑/##tRelevante·nuancec(OpenFTV)o fork · x to clear history · Esc to close│
                                                                           │
● HoFijnmazigeepoliciesh(bijv.s`resource.status ==