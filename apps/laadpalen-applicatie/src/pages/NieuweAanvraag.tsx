import {useState} from "react";
import {useNavigate} from "react-router-dom";
import {Header} from "@/components/Header";
import {Card} from "@/components/ui/card";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {ApplicationForm} from "@/components/ApplicationForm";
import {ArrowLeft} from "lucide-react";
import {format} from "date-fns";
import {nl} from "date-fns/locale";
import {useAanvragen, useCreateAanvraag, useGebruikers} from "@/lib/api";
import {useCurrentUser} from "@/contexts/CurrentUserContext";

interface AuditLogEntry {
  id: string;
  timestamp: Date;
  action: string;
  system: string;
  result: "success" | "error" | "info";
  details: string;
  isPublic?: boolean;
}

const NieuweAanvraag = () => {
  const navigate = useNavigate();
  const [auditLog, setAuditLog] = useState<AuditLogEntry[]>([]);
  // The API stores an aanvraag under the id the client picks, so take the next free one.
  const {data: aanvragen} = useAanvragen();
  const applicationId = Math.max(0, ...(aanvragen ?? []).map((a) => a.id)) + 1;
  const {currentUser, setCurrentUser} = useCurrentUser();
  const createAanvraag = useCreateAanvraag();
  const {data: gebruikers = []} = useGebruikers();

  const addAuditEntry = (entry: Omit<AuditLogEntry, "id" | "timestamp">) => {
    const newEntry: AuditLogEntry = {
      ...entry,
      id: crypto.randomUUID(),
      timestamp: new Date(),
    };
    setAuditLog((prev) => [newEntry, ...prev]);
  };

  const handleSubmit = (data: { postcode: string; huisnummer: number; }) => {
    if (!aanvragen) {
      addAuditEntry({
        action: "Aanvraag opslaan mislukt",
        system: "API",
        result: "error",
        details: "De bestaande aanvragen zijn nog niet geladen, probeer het opnieuw",
        isPublic: true,
      });
      return;
    }

    const statusDatum = new Date().toISOString();

    createAanvraag.mutate(
      {
        id: applicationId.toString(),
        data: {
          postcode: data.postcode,
          huisnummer: data.huisnummer,
          status: 1,
          naam: "",
          omschrijving: "",
          "status-datum": statusDatum,
          id: applicationId
        },
      },
      {
        onSuccess: () => {
          navigate(`/behandelaar/aanvraag/${applicationId}`);
        },
        onError: (error) => {
          console.error("Failed to create aanvraag:", error);
          addAuditEntry({
            action: "Aanvraag opslaan mislukt",
            system: "API",
            result: "error",
            details: error instanceof Error ? error.message : "Er is een fout opgetreden bij het opslaan van de aanvraag",
            isPublic: true,
          });
        },
      }
    );
  };

  return (
    <div className="min-h-screen bg-background">
      <Header
        currentUser={currentUser}
        onUserChange={setCurrentUser}
      />

      <main className="container mx-auto px-4 py-8 max-w-7xl">
        <Button
          variant="outline"
          onClick={() => navigate("/behandelaar")}
          className="mb-6"
        >
          <ArrowLeft className="mr-2 h-4 w-4"/>
          Terug naar overzicht
        </Button>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <Card className="p-8 shadow-lg">
              <div className="mb-6">
                <h1 className="text-3xl font-bold text-foreground mb-2">
                  Nieuwe Aanvraag
                </h1>
                <p className="text-muted-foreground">
                  Vul de gegevens in om een nieuwe laadpaalvergunning aanvraag aan te maken.
                </p>
              </div>
              <ApplicationForm
                onSubmit={handleSubmit}
                onAuditLog={addAuditEntry}
                currentUser={currentUser}
                gebruikers={gebruikers}
              />
              {createAanvraag.isPending && (
                <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <p className="text-sm text-blue-800">Aanvraag wordt opgeslagen...</p>
                </div>
              )}
            </Card>
          </div>

          <div className="lg:col-span-1">
            <Card
              className="p-6 sticky top-6 max-h-[calc(100vh-3rem)] flex flex-col rounded-none bg-transparent border-0">
              <p className="text-sm text-muted-foreground mb-4 flex-shrink-0">
                Chronologisch overzicht van feedback en beslissingen
              </p>

              {/* Reserved space for sticky item - empty for this page */}
              <div className="mb-4 flex-shrink-0">
                {/* No sticky item on this page */}
              </div>

              <div className="space-y-3 overflow-y-auto flex-1 relative pl-8">
                {/* Vertical timeline line */}
                <div className="absolute left-2 top-0 bottom-0 w-px bg-border"></div>
                {auditLog.length === 0 ? (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    Geen acties gelogd
                  </p>
                ) : (
                  auditLog.map((entry) => (
                    <div key={entry.id} className="relative">
                      {/* Timeline dot */}
                      <div
                        className="absolute -left-8 top-3 w-3 h-3 rounded-full bg-primary border-2 border-background z-10"></div>
                      {/* Connection line */}
                      <div className="absolute -left-7 top-4 w-6 h-px bg-border"></div>

                      <div
                        className={`border rounded-none p-3 text-sm ${
                          entry.isPublic
                            ? 'border-primary/50 bg-primary/5'
                            : 'border-border'
                        }`}
                      >
                        <div className="flex items-start justify-between mb-2">
                          <div className="flex items-center gap-2">
                            <Badge
                              variant="outline"
                              className={
                                entry.result === "success"
                                  ? "border-success text-success"
                                  : entry.result === "error"
                                    ? "border-destructive text-destructive"
                                    : "border-muted-foreground text-muted-foreground"
                              }
                            >
                              {entry.system}
                            </Badge>
                            {entry.isPublic && (
                              <Badge variant="secondary" className="text-xs">
                                Openbaar
                              </Badge>
                            )}
                          </div>
                          <span className="text-xs text-muted-foreground">
                          {format(entry.timestamp, "HH:mm:ss", {locale: nl})}
                        </span>
                        </div>
                        <p className="font-medium text-foreground mb-1">
                          {entry.action}
                        </p>
                        <p className="text-muted-foreground text-xs">
                          {entry.details}
                        </p>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
};

export default NieuweAanvraag;
