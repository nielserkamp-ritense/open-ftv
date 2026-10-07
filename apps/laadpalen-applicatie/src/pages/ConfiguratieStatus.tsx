import {useNavigate} from "react-router-dom";
import {useQueries, useQuery} from "@tanstack/react-query";
import {Header} from "@/components/Header";
import {useCurrentUser} from "@/contexts/CurrentUserContext";
import {Card} from "@/components/ui/card";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {ArrowLeft, CheckCircle2, XCircle, Loader2, Database, Users, Car, Info} from "lucide-react";
import {
  apiClient, brpApiClient, rdwApiClient,
  vlierdamPdp1Client,
  evaluateLaadpaalIndienen, newTraceId,
  ApiError, isApiError, reattributeChainError,
} from "@/lib/api";
import {VITE_API_BASE_URL, VITE_VLIERDAM_OUTWAY_URL} from "@/lib/env";

interface MetaTable {
  id: string;
  description?: string;
}

interface MetaDatasource {
  id: string;
  description?: string;
  tables?: MetaTable[];
}

interface HealthCheck {
  key: string;
  // chain hops; source must match the ApiError source thrown for that hop to highlight it.
  segments: { text: string; source?: string }[];
  queryFn: () => Promise<unknown>;
}

interface SourceConfig {
  key: string;
  title: string;
  icon: React.ReactNode;
  url: string;
  client: { get<T>(endpoint: string): Promise<T> };
  datasourceId: string;
  healthChecks: HealthCheck[];
}

const StatusBadge = ({isLoading, isError}: { isLoading: boolean; isError: boolean }) => {
  if (isLoading) {
    return (
      <Badge variant="outline" className="gap-1">
        <Loader2 className="h-3.5 w-3.5 animate-spin"/>
        Controleren...
      </Badge>
    );
  }
  if (isError) {
    return (
      <Badge className="bg-destructive text-destructive-foreground hover:bg-destructive gap-1">
        <XCircle className="h-3.5 w-3.5"/>
        Niet bereikbaar
      </Badge>
    );
  }
  return (
    <Badge className="bg-success text-success-foreground hover:bg-success gap-1">
      <CheckCircle2 className="h-3.5 w-3.5"/>
      Online
    </Badge>
  );
};

const SourceCard = ({source}: { source: SourceConfig }) => {
  const {data, isLoading, isError} = useQuery({
    queryKey: ["meta", "datasource", source.key],
    queryFn: () => source.client.get<MetaDatasource>(`/v1/meta/datasource/${source.datasourceId}`),
    retry: false,
    refetchOnWindowFocus: false,
  });

  const healthResults = useQueries({
    queries: source.healthChecks.map((check) => ({
      queryKey: ["meta", "health", source.key, check.key],
      queryFn: check.queryFn,
      retry: false,
      refetchOnWindowFocus: false,
    })),
  });

  const tables = data?.tables;
  const anyLoading = isLoading || healthResults.some((r) => r.isLoading);
  const anyError = isError || healthResults.some((r) => r.isError);

  return (
    <Card className="p-6 space-y-4">
      <div className="flex items-center gap-2">
        {source.icon}
        <h3 className="font-semibold text-foreground">{source.title}</h3>
      </div>

      <div className="space-y-1">
        <p className="text-xs text-muted-foreground">URL</p>
        <p className="text-sm font-mono break-all">{source.url}</p>
      </div>

      <div className="space-y-1">
        <p className="text-xs text-muted-foreground">Status</p>
        <StatusBadge isLoading={anyLoading} isError={anyError}/>
      </div>

      {source.healthChecks.length > 0 && (
        <div className="space-y-1">
          <p className="text-xs text-muted-foreground">Onderliggende diensten</p>
          <ul className="text-sm space-y-0.5">
            {source.healthChecks.map((check, i) => {
              const result = healthResults[i];
              const failedSource = result.isError && isApiError(result.error) ? result.error.source : undefined;
              return (
                <li key={check.key} className="flex items-center gap-1.5">
                  {result.isLoading ? <Loader2 className="h-3 w-3 animate-spin text-muted-foreground"/> : result.isError ?
                    <XCircle className="h-3 w-3 text-destructive"/> :
                    <CheckCircle2 className="h-3 w-3 text-success"/>}
                  <span>
                    Keten (
                    {check.segments.map((segment, si) => (
                      <span key={segment.text}>
                        {si > 0 && " → "}
                        <span className={segment.source && segment.source === failedSource ? "text-destructive font-medium" : undefined}>
                          {segment.text}
                        </span>
                      </span>
                    ))}
                    )
                  </span>
                </li>
              );
            })}
          </ul>
        </div>
      )}

      <div className="space-y-1">
        <p className="text-xs text-muted-foreground">Beschikbare gegevens</p>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Laden...</p>
        ) : tables && tables.length > 0 ? (
          <ul className="text-sm space-y-0.5 list-disc list-inside">
            {tables.map((table) => (
              <li key={table.id}>
                <span className="font-mono">{table.id}</span>
                {table.description ? ` — ${table.description}` : ""}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">Geen gegevens beschikbaar</p>
        )}
      </div>
    </Card>
  );
};

// The full intended chain stays visible (matches the real infra, see project memory), but
// only the hop a real laadpalen aanvraag actually touches gets pinged for real - the other
// segments are shown for context only and never highlighted, since they're never called.
const vlierdamHealthChecks: HealthCheck[] =
  vlierdamPdp1Client ? [{
    key: "pdp-outway",
    segments: [{text: "pdp1", source: "openftv-vlierdam-pdp"}, {text: "outway"}],
    queryFn: async () => {
      try {
        await evaluateLaadpaalIndienen({gebruiker: "healthcheck", postcode: "0000AA", huisnummer: 0, traceId: newTraceId()});
      } catch (err) {
        throw isApiError(err) ? new ApiError(err.message, err.status, err.data, "openftv-vlierdam-pdp") : err;
      }
      return true;
    },
  }] : [];

// Builds a single health check that pings `path` via `client`, labeled with the three chain segments for display.
const dataspaceChainCheck = (
  client: { post<T>(endpoint: string, body?: unknown): Promise<T> },
  path: string,
  inwaySource: string,
  pdpSource: string,
  dataspaceSource: string,
  body?: unknown,
): HealthCheck[] => [{
  key: "chain",
  segments: [{text: "inway", source: inwaySource}, {text: "pdp", source: pdpSource}, {text: "dataspace", source: dataspaceSource}],
  queryFn: async () => {
    try {
      return await client.post(path, body);
    } catch (err) {
      reattributeChainError(err, {inway: inwaySource, pdp: pdpSource});
    }
  },
}];

const sources: SourceConfig[] = [
  {
    key: "vlierdam",
    title: "Gemeente Vlierdam - Laadpalen",
    icon: <Database className="h-5 w-5 text-primary"/>,
    url: VITE_API_BASE_URL,
    client: apiClient,
    datasourceId: "vlierdam",
    healthChecks: vlierdamHealthChecks,
  },
  {
    key: "brp",
    title: "BRP",
    icon: <Users className="h-5 w-5 text-primary"/>,
    url: VITE_VLIERDAM_OUTWAY_URL,
    client: brpApiClient,
    datasourceId: "brp",
    // Real end-to-end ping - client points at the Vlierdam outway, gated by AuthZEN at both
    // the outway and the RvIG inway (see VITE_VLIERDAM_OUTWAY_URL/VITE_BRP_API_KEY).
    healthChecks: dataspaceChainCheck(
      brpApiClient, "/v1/haalcentraal/api/brp/personen", "openftv-rvig-inway", "openftv-rvig-pdp1", "openftv-rvig-dataspace",
    ),
  },
  {
    key: "brv",
    title: "BRV (RDW)",
    icon: <Car className="h-5 w-5 text-primary"/>,
    url: VITE_VLIERDAM_OUTWAY_URL,
    client: rdwApiClient,
    datasourceId: "brv",
    healthChecks: dataspaceChainCheck(
      rdwApiClient, "/v1/voertuigen", "openftv-rdw-inway", "openftv-rdw-pdp1", "openftv-rdw-dataspace",
      // A non-diplomatic owner: RDW refuses a diplomatic vehicle, or in scenario 2 the outway does without a
      // Gemeentesecretaris.
      {"@filter": "kenteken.bsn=999990263"},
    ),
  },
];

const ConfiguratieStatus = () => {
  const navigate = useNavigate();
  const {currentUser, setCurrentUser} = useCurrentUser();

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

        <div className="mb-8">
          <h1 className="text-4xl font-bold text-foreground mb-2">
            Config &amp; Status Bronnen
          </h1>
          <p className="text-lg text-muted-foreground">
            Overzicht van de gekoppelde bronnen: waar deze applicatie mee praat, en of ze bereikbaar zijn
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {sources.map((source) => (
            <SourceCard key={source.key} source={source}/>
          ))}
        </div>

        <Card className="p-6 mt-6 bg-muted/30">
          <div className="flex items-start gap-3">
            <Info className="h-5 w-5 text-muted-foreground mt-0.5"/>
            <div className="text-sm text-muted-foreground">
              <p className="font-medium text-foreground mb-1">Over deze gegevens</p>
              <p>
                Alle drie de bronnen zijn testdata, geladen in het geheugen bij het opstarten.
              </p>
            </div>
          </div>
        </Card>
      </main>
    </div>
  );
};

export default ConfiguratieStatus;
