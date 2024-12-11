# policy voor het pro-actief berekenen van zorgtoeslag.
# wordt uitgevoerd door de FSC Outway van de Dienst Zorgtoeslagen,
# op basis van het doel dat in het request meegegeven wordt.

package doelbinding.zorgtoeslag

import rego.v1

default allow := false

# bepaal welke service aangeroepen wordt.
# dit halen we op uit externe data (PIP),
# met een key die samengesteld is met de URL van de FSC Inway + de naam van het FSC contract (request).

service := data.entities.service[input.uri]

# bepaal of beide partijen lid zijn van het FDS stelsel.
# dit doen we met de FDS ledenlijst (via de PIP opgehaald),
# en het oin van de partijen (uit het FSC contract).

lid1 := data.members[input.source.oin]
lid2 := data.members[input.target.oin]

allow if {
    # voor deze doelbinding is alleen een GET toegestaan op het /networth path.

    input.http.path == "/networth"
    input.http.method == "GET"

    # beide partijen zijn lid.

    lid1 == true
    lid2 == true

    # de service die bevraagd wordt, wordt door de Belastingdienst aangeboden.

    service.code == "backend"
    service.owner == "Belastingdienst"
}
