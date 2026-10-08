import {Link} from "react-router-dom";
import { ArrowRight, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {VITE_BEHEER_URL, VITE_RDW_BEHEER_URL, VITE_RVIG_BEHEER_URL} from "@/lib/env.ts";

const Index = () => {
  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8 flex items-center justify-center min-h-[calc(100vh-120px)]">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="flex flex-col gap-8 w-full max-w-4xl col-span-2">
              <Card className="p-8">
                <div className="text-xl font-bold mb-4">
                  Demo applicatie OpenFTV
                </div>
                <div className="space-y-4">
                  <p className="text-lg leading-relaxed">
                    Dit is de demo-omgeving van de FTV-applicatie voor de fictieve gemeente Vlierdam. In deze demo zie je hoe medewerkers laadpaalvergunningen aanvragen en daarbij alleen de strikt noodzakelijke gegevens uit BRP, BRV en het zaaksysteem mogen gebruiken.
                  </p>
                  <p className="text-lg leading-relaxed">
                    De applicatie bestaat uit een zaaksysteemmodule (front-end) voor het aanvragen en beoordelen van vergunningen en een beheeromgeving (OpenFTV) waarin je de toegangsregels kunt inzien en aanpassen. Gebruik een postcode die begint met 1111 om de <Link className={'underline'} to={"https://vng-realisatie.github.io/ftv/toepassen/demonstratie/"}>testscenario's</Link> te doorlopen en verschillende uitkomsten te zien.
                  </p>
                  <p className="text-lg leading-relaxed">
                    Op de <Link className={'underline'} to={"https://vng-realisatie.github.io/ftv/"}>FTV-site</Link> lees je stap voor stap hoe de case is opgebouwd en welke scenario's je kunt proberen. Heb je vragen of feedback? Laat het weten via <Link  className={'underline'} to={"https://digilab.overheid.nl/chat/digilab/channels/federatieve-toegangsverlening"}>Mattermost</Link>.
                  </p>
                </div>
                <div className="my-8 space-y-6">
                  <Button asChild size="lg" className="w-full h-14 text-lg font-semibold">
                    <a href="/behandelaar" target="_blank" aria-describedby="nieuw-tabblad">
                      Open het Zaaksysteem Laadpalen
                      <ArrowRight />
                    </a>
                  </Button>

                  <div className="space-y-3">
                    <div className="text-sm font-medium text-muted-foreground">Beheeromgevingen</div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      {[
                        {organisation: "Gemeente Vlierdam", url: VITE_BEHEER_URL},
                        {organisation: "RDW", url: VITE_RDW_BEHEER_URL},
                        {organisation: "RvIG", url: VITE_RVIG_BEHEER_URL},
                      ].map(({organisation, url}, index) => (
                        <a key={organisation}
                           href={url}
                           target="_blank"
                           aria-labelledby={`beheer-${index}`}
                           aria-describedby={`beheer-${index}-omschrijving nieuw-tabblad`}
                           className="group flex items-center justify-between gap-4 rounded-md border p-4 transition-colors hover:border-primary hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        >
                          <div>
                            <div id={`beheer-${index}`} className="font-semibold">{organisation}</div>
                            <div id={`beheer-${index}-omschrijving`} className="text-sm text-muted-foreground">Toegangsregels beheren</div>
                          </div>
                          <ExternalLink className="size-4 shrink-0 text-muted-foreground group-hover:text-primary" />
                        </a>
                      ))}
                    </div>
                  </div>
                  <span id="nieuw-tabblad" className="sr-only">opent in een nieuw tabblad</span>
                </div>
                <div className={"flex justify-center mt-4"}>
                  <Link to={"https://vng-realisatie.github.io/ftv/"}><img src={"logo_tekst.svg"} className={"h-[60px] opacity-50"}></img></Link>
                </div>
              </Card>
            </div>
        </div>
      </main>
    </div>
  );
};

export default Index;
