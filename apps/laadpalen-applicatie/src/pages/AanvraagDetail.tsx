/* eslint-disable @typescript-eslint/no-explicit-any */
import {useState, useEffect, useMemo} from "react";
import {useParams, useNavigate} from "react-router-dom";
import {Header} from "@/components/Header";
import {Card} from "@/components/ui/card";
import {Button} from "@/components/ui/button";
import {Badge} from "@/components/ui/badge";
import {Textarea} from "@/components/ui/textarea";
import {Label} from "@/components/ui/label";
import {Input} from "@/components/ui/input";
import {useCurrentUser} from "@/contexts/CurrentUserContext";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {AddressMap} from "@/components/AddressMap";
import {
  CheckCircle2,
  XCircle,
  Clock,
  ArrowLeft,
  Database,
  Car,
  UserCheck,
  FileText,
  Send,
  AlertCircle,
  Loader2,
  Check,
  MapPin,
  Info,
} from "lucide-react";
import {format} from "date-fns";
import {nl} from "date-fns/locale";
import houseFront1 from "@/assets/house-front-1.jpg";
import houseFront2 from "@/assets/house-front-2.jpg";
import houseFront3 from "@/assets/house-front-3.jpg";
import {
  Aanvraag,
  type BrpPerson,
  RdwVehicle,
  useAanvragen,
  useBrpAddress,
  useRdwVehiclesForMultipleBsns,
  useLaadpalen,
  usePatchAanvraag,
  useGebruikers,
  useDiplomaatBeslissing,
  getDutchErrorMessage,
  getErrorMessage
} from "@/lib/api";

interface AuditLogEntry {
  id: string;
  timestamp: Date;
  action: string;
  system: "BRP" | "BRV" | "GLP" | "FTV" | "FSC" | "FDS" | "LDV" | "LTV" | "Zaaksysteem" | "Ambtenaar";
  connectionType?: "FSC" | "Internal"; // FSC voor externe systemen, Internal voor gemeentelijk
  accessControl?: "FTV"; // Voor toegangscontrole
  logging?: "LDV" | "LTV"; // LDV voor dataverwerking, LTV voor toegangsverlening
  result: "success" | "error" | "info";
  details: string;
  isPublic?: boolean; // true = openbare feedback, false/undefined = technisch

  // AuthZEN specifieke velden (voor FTV toegangscontrole logs)
  authzenDecision?: "permit" | "deny";
  authzenReason?: string;

  // Data response velden (voor bronsysteem data logs)
  dataResponse?: {
    recordsFound?: number;
    summary?: string;
  };
}

// Mapping voor Vlierdam adressen
const vlierdamStreets: Record<string, string> = {
  "1111AA": "Korenbloemplein",
  "1111AB": "Korenbloemplein",
  "1111BA": "Wilgenlaan",
  "1111BB": "Wilgenlaan",
  "1111CA": "Esdoornstraat",
  "1111CB": "Esdoornstraat",
  "1111DA": "Populierweg",
  "1111DB": "Populierweg",
  "1112AA": "Lindenlaan",
  "1112AB": "Lindenlaan",
};

const getFullAddress = (postcode: string, huisnummer: number | string) => {
  const straat = vlierdamStreets[postcode] || "Hoofdstraat";
  return `${straat} ${huisnummer}, ${postcode} Vlierdam`;
};

const getHouseImage = (id: number | string) => {
  const images = [houseFront1, houseFront2, houseFront3];
  const index = (typeof id === 'number' ? id : parseInt(id, 10)) % images.length;
  return images[index];
};

// BRP dates arrive as "YYYYMMDD" (e.g. "19541216") - format as "1954-12-16".
const formatGeboortedatum = (raw?: string) => {
  if (!raw || raw.length !== 8) return raw;
  return `${raw.slice(0, 4)}-${raw.slice(4, 6)}-${raw.slice(6, 8)}`;
};

const AanvraagDetail = () => {
  const {id} = useParams<{ id: string }>();
  const navigate = useNavigate();
  const {data: aanvragenData, isLoading, isFetching, error} = useAanvragen();
  const application = useMemo(
    () => aanvragenData?.find((app) => app.id.toString() === id) ?? null,
    [aanvragenData, id]
  );
  const [toelichting, setToelichting] = useState("");
  const [zaaknummer, setZaaknummer] = useState("");
  const [isCheckingBRP, setIsCheckingBRP] = useState(false);
  const [isCheckingRDW, setIsCheckingRDW] = useState(false);
  const [isSendingToZaak, setIsSendingToZaak] = useState(false);
  const [showAuditLog, setShowAuditLog] = useState(true);
  const {currentUser, setCurrentUser} = useCurrentUser();

  const [brpData, setBrpData] = useState<BrpPerson[]>(null);
  const [rdwData, setRdwData] = useState<RdwVehicle[]>(null);
  const [existingLaadpaalData, setExistingLaadpaalData] = useState<any>(null);
  const [isCheckingLaadpaal, setIsCheckingLaadpaal] = useState(false);
  const [vergunningBeslissing, setVergunningBeslissing] = useState<"goedkeuren" | "afwijzen" | null>(null);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [isVerzonden, setIsVerzonden] = useState(false);
  const [selectedLogEntry, setSelectedLogEntry] = useState<AuditLogEntry | null>(null);
  const [enableBrpCheck, setEnableBrpCheck] = useState(false);
  const [enableRdwCheck, setEnableRdwCheck] = useState(false);
  const [enableLaadpalenCheck, setEnableLaadpalenCheck] = useState(false);
  const [updateError, setUpdateError] = useState<string | null>(null);

  // Check user role - decided by Vlierdam's own PDP2 (Rego), gated on whether this aanvraag
  // involves a diplomatic vehicle and the acting employee's afdeling.
  const { data: gebruikersData } = useGebruikers();
  const huidigeGebruiker = gebruikersData?.find(g => g.naam === currentUser);
  const involvesDiplomaticVehicle = !!rdwData?.some((v: RdwVehicle) => v.diplomatiek);
  const diplomaatQuery = useDiplomaatBeslissing({
    aanvraagId: application?.id,
    gebruiker: currentUser,
    afdeling: huidigeGebruiker?.afdeling,
    involvesDiplomaticVehicle,
    enabled: !!application && !!huidigeGebruiker,
  });
  const diplomaatBeslissing = diplomaatQuery.data;
  const isGemeentesecretaris = diplomaatBeslissing?.decision ?? false;

  // Fail closed: only an explicit decision:true from pdp2 lifts the block. Still loading or a
  // failed check must not silently behave as "allowed" - both keep blocking until resolved.
  const diplomaatCheckPending = involvesDiplomaticVehicle && (diplomaatQuery.isLoading || diplomaatQuery.isFetching);
  const diplomaatCheckFailed = involvesDiplomaticVehicle && diplomaatQuery.isError;
  const requiresAdditionalAuthorization = involvesDiplomaticVehicle && diplomaatBeslissing?.decision !== true;

  // Mutation hook for updating aanvraag
  const updateAanvraagMutation = usePatchAanvraag();

  // BRP query hook - enabled only when check is triggered
  const brpQuery = useBrpAddress(
    application?.postcode || '',
    application?.huisnummer || 0,
    { enabled: enableBrpCheck && !!application }
  );

  // RDW query hook - enabled only when check is triggered and brpData is available
  const bsnList = brpData?.map(b => b.bsn).filter(Boolean) || [];
  const rdwQuery = useRdwVehiclesForMultipleBsns(bsnList, {
    enabled: enableRdwCheck && bsnList.length > 0
  });

  // Laadpalen query hook - enabled only when check is triggered
  const laadpalenQuery = useLaadpalen(
    application ? {
      "@filter":  "laadpaal.postcode=" + application.postcode + ",laadpaal.huisnummer=" + application.huisnummer
    } : undefined,
    { enabled: enableLaadpalenCheck && !!application }
  );

  // Handle BRP query state changes
  useEffect(() => {
    if (!enableBrpCheck) return;

    if (brpQuery.isLoading) {
      setIsCheckingBRP(true);
    } else if (brpQuery.isSuccess) {
      setIsCheckingBRP(false);
      const data = brpQuery.data || [];
      setBrpData(data);

      if (data.length === 0) {
        console.error("BRP address data is empty");
      }
    } else if (brpQuery.isError) {
      setIsCheckingBRP(false);
      console.error("BRP check failed:", brpQuery.error);
    }
  }, [brpQuery.isLoading, brpQuery.isSuccess, brpQuery.isError, brpQuery.error, brpQuery.data, enableBrpCheck]);

  // Handle RDW query state changes
  useEffect(() => {
    if (!enableRdwCheck) return;
    if (rdwQuery.isLoading) {
      setIsCheckingRDW(true);
    } else if (rdwQuery.isSuccess) {
      setIsCheckingRDW(false);
      const data = rdwQuery.data;
      setRdwData(data);

      const elektrischeVoertuigen = data.filter(v => v.klasse === "E")

      const auditDetails = elektrischeVoertuigen.length === 0
        ? "Geen elektrische voertuigen geregistreerd op naam van bewoner(s)"
        : `${elektrischeVoertuigen.length} elektrisch voertuig(en) gevonden: ${elektrischeVoertuigen.map(v => `${v.merk} (${v.kenteken})`).join(", ")}`;
    } else if (rdwQuery.isError) {
      setIsCheckingRDW(false);
      console.error("RDW check failed:", rdwQuery.error);
    }
  }, [rdwQuery.isLoading, rdwQuery.isSuccess, rdwQuery.isError, rdwQuery.error, rdwQuery.data, enableRdwCheck]);

  // Handle Laadpalen query state changes
  useEffect(() => {
    if (!enableLaadpalenCheck) return;

    if (laadpalenQuery.isLoading) {
      setIsCheckingLaadpaal(true);
    } else if (laadpalenQuery.isSuccess) {
      setIsCheckingLaadpaal(false);
      const data = laadpalenQuery.data;
      const bestaandeLaadpaal = data && data.length > 0;

      console.log("Laadpalen data:", data);
      setExistingLaadpaalData({
        bestaandeLaadpaal,
        adres: application ? getFullAddress(application.postcode, application.huisnummer) : '',
      });
    } else if (laadpalenQuery.isError) {
      setIsCheckingLaadpaal(false);
      console.error("Laadpaal check failed:", laadpalenQuery.error);
    }
  }, [laadpalenQuery.isLoading, laadpalenQuery.isSuccess, laadpalenQuery.isError, laadpalenQuery.error, laadpalenQuery.data, enableLaadpalenCheck, application]);

  useEffect(() => {
    if (application) {
      setZaaknummer(`ZAAK-2025-${application.id.toString().padStart(4, '0')}`);
    }
  }, [application]);

  const handleBRPCheck = async () => {
    if (!application) return;
    if (enableBrpCheck) {
      brpQuery.refetch();
    } else {
      setEnableBrpCheck(true);
    }
  };

  const handleRDWCheck = async () => {
    if (!brpData) {
      return;
    }
    if (enableRdwCheck) {
      rdwQuery.refetch();
    } else {
      setEnableRdwCheck(true);
    }
  };

  const handleCheckLaadpaal = async () => {
    if (!application) return;
    if (enableLaadpalenCheck) {
      laadpalenQuery.refetch();
    } else {
      setEnableLaadpalenCheck(true);
      setIsCheckingLaadpaal(true);
    }
  };

  const handleBehandelingAfronden = () => {
    if (!application) return;
    console.log("Behandeling afronden geklikt voor aanvraag:", application.id);
    if (!vergunningBeslissing) {
      return;
    }

    // Check of er een mismatch is tussen beslissing en controleresultaten
    const hasElectricVehicle = rdwData?.some((b: RdwVehicle) => b.klasse === "E");
    const noExistingLaadpaal = existingLaadpaalData && !existingLaadpaalData.bestaandeLaadpaal;
    const allChecksCompleted = brpData != undefined && rdwData != undefined && existingLaadpaalData != undefined;
    const canApprove = hasElectricVehicle && noExistingLaadpaal;

    const hasMismatch = allChecksCompleted && (
      (vergunningBeslissing === "goedkeuren" && !canApprove) ||
      (vergunningBeslissing === "afwijzen" && canApprove)
    );

    // Toelichting is alleen verplicht bij mismatch
    if (hasMismatch && !toelichting.trim()) {
      return;
    }

    // Toon bevestigingsdialoog
    setShowConfirmDialog(true);
  };

  const confirmSendToZaaksysteem = async () => {
    if (!application || !vergunningBeslissing) return;

    setShowConfirmDialog(false);
    setIsSendingToZaak(true);
    setUpdateError(null);

    try {
      // Determine the new status based on decision
      const newStatus = vergunningBeslissing === "goedkeuren" ? 4 : 8;

      // Call backend to update aanvraag
      const response = await updateAanvraagMutation.mutateAsync({
        id: application.id.toString(),
        data: {
          status: newStatus,
          "afgekeurd-reden": toelichting || undefined,
        },
      });

      if (response) {
        console.log("Aanvraag succesvol bijgewerkt:", response);
      }

      setIsSendingToZaak(false);
      setIsVerzonden(true);

      // Navigate back to behandelaar overview
      navigate("/behandelaar");
    } catch (error) {
      setIsSendingToZaak(false);
      setUpdateError(error instanceof Error ? error.message : "Er is een fout opgetreden bij het verzenden naar het zaaksysteem");
      console.error("Failed to update aanvraag:", error);
    }
  };

  const getStatusBadge = (status: Aanvraag["status"]) => {
    switch (status) {
      case 4:
        return (
          <Badge className="bg-success text-success-foreground">
            <CheckCircle2 className="mr-1 h-4 w-4"/>
            Toegewezen
          </Badge>
        );
      case 8:
        return (
          <Badge className="bg-destructive text-destructive-foreground">
            <XCircle className="mr-1 h-4 w-4"/>
            Afgewezen
          </Badge>
        );
      case 2:
        return (
          <Badge className="bg-blue-500 text-white">
            <Clock className="mr-1 h-4 w-4"/>
            In behandeling
          </Badge>
        );
      case 1:
        return (
          <Badge className="bg-warning text-warning-foreground">
            <Clock className="mr-1 h-4 w-4"/>
            Te behandelen
          </Badge>
        );
    }
  };

  if (!application && (isLoading || isFetching)) {
    return (
      <div className="min-h-screen bg-background">
        <Header currentUser={currentUser} onUserChange={setCurrentUser} showAuditLog={showAuditLog} onToggleAuditLog={() => setShowAuditLog(!showAuditLog)}/>
        <div className="container mx-auto px-4 py-8 text-center">
          <Loader2 className="h-16 w-16 mx-auto text-muted-foreground mb-4 animate-spin"/>
          <h2 className="text-2xl font-bold mb-2">Aanvraag wordt geladen...</h2>
        </div>
      </div>
    );
  }

  if (!application) {
    return (
      <div className="min-h-screen bg-background">
        <Header currentUser={currentUser} onUserChange={setCurrentUser} showAuditLog={showAuditLog} onToggleAuditLog={() => setShowAuditLog(!showAuditLog)}/>
        <div className="container mx-auto px-4 py-8 text-center">
          <AlertCircle className="h-16 w-16 mx-auto text-destructive mb-4"/>
          <h2 className="text-2xl font-bold mb-2">Aanvraag niet gevonden</h2>
          <Button onClick={() => navigate("/behandelaar")}>
            <ArrowLeft className="mr-2 h-4 w-4"/>
            Terug naar overzicht
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Header
        currentUser={currentUser}
        onUserChange={setCurrentUser}
        showAuditLog={showAuditLog}
        onToggleAuditLog={() => setShowAuditLog(!showAuditLog)}
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
          {/* Main content - 2 columns */}
          <div className="lg:col-span-2 space-y-6">
            {/* Application Details */}
            <Card className="p-6">
              <div className="flex items-start justify-between mb-6">
                <div>
                  <h1 className="text-3xl font-bold text-foreground mb-2 flex items-center gap-2">
                    <MapPin className="h-7 w-7 text-primary"/>
                    {getFullAddress(application.postcode, application.huisnummer)}
                  </h1>
                  <p className="text-muted-foreground">
                    Aanvraag #{application.id}
                  </p>
                </div>
                {getStatusBadge(application.status)}
              </div>

              <div className="grid grid-cols-2 gap-6 mb-6">
                <div>
                  <Label className="text-muted-foreground">Aanvraagdatum</Label>
                  <p className="text-lg font-medium text-foreground">
                    {format(application["status-datum"], "d MMMM yyyy", {locale: nl})}
                  </p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Zaaknummer</Label>
                  <p className="text-lg font-medium text-foreground font-mono">
                    {zaaknummer}
                  </p>
                </div>
              </div>

              {/* Vooraanzicht pand */}
              <div className="mb-4">
                <Label className="text-muted-foreground mb-2 block">Vooraanzicht pand</Label>
                <img
                  src={getHouseImage(application.id)}
                  alt={`Vooraanzicht ${getFullAddress(application.postcode, application.huisnummer)}`}
                  className="w-full h-64 object-cover rounded-lg border border-border"
                />
              </div>

              {/* Kaart met locatie laadpaal */}
              <div>
                <Label className="text-muted-foreground mb-2 block">Locatie laadpaal</Label>
                <AddressMap
                  address={getFullAddress(application.postcode, application.huisnummer)}
                  postcode={application.postcode}
                />
              </div>
            </Card>

            {/* External System Checks */}
            <Card className="p-6">
              <h2 className="text-xl font-bold text-foreground mb-4 flex items-center gap-2">
                <Database className="h-5 w-5"/>
                Externe Controles (via FTV)
              </h2>
              <p className="text-sm text-muted-foreground mb-6">
                Voer controles uit bij BRP en RDW via het Federatief Toegangsstelsel voor Vertrouwde dienstverlening
              </p>

              <div className="space-y-4">
                {/* BRP Check */}
                <div className="border border-border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <UserCheck className="h-5 w-5 text-primary"/>
                      <h3 className="font-semibold text-foreground">BRP Controle</h3>
                    </div>
                    <Button
                      onClick={handleBRPCheck}
                      disabled={isCheckingBRP || !!brpData}
                      size="sm"
                    >
                      {isCheckingBRP ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                          Controleren...
                        </>
                      ) : brpData ? (
                        <>
                          <Check className="mr-2 h-4 w-4"/>
                          Voltooid
                        </>
                      ) : (
                        "Uitvoeren"
                      )}
                    </Button>
                  </div>

                  {brpData && (
                    <div className="bg-muted/50 rounded p-3 space-y-3 text-sm">
                      <div className="font-medium text-foreground mb-2">
                        {brpData.length} {brpData.length === 1 ? 'bewoner' : 'bewoners'} gevonden
                      </div>
                      {brpData.map((bewoner: BrpPerson, index: number) => (
                        <div key={index} className="pb-3 border-b last:border-b-0 last:pb-0">
                          <div className="grid grid-cols-2 gap-2">
                            <div>
                              <span className="text-muted-foreground">BSN:</span>
                              <span className="ml-2 font-mono">{bewoner.bsn}</span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Naam:</span>
                              <span className="ml-2">{bewoner.voornamen} {bewoner.geslachtsnaam}</span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Gemeente:</span>
                              <span className="ml-2">{bewoner.gemeente}</span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Geboortedatum:</span>
                              <span className="ml-2">{formatGeboortedatum(bewoner.geboortedatum)}</span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Meerderjarig:</span>
                              <span className="ml-2">{bewoner.volwassen ? "Ja" : "Nee"}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}

                  {brpQuery.isError && (
                    <div className="bg-destructive/10 border border-destructive/30 rounded p-3 text-sm">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-5 w-5 text-destructive mt-0.5 shrink-0"/>
                        <div>
                          <p className="font-medium">{getDutchErrorMessage(brpQuery.error)}</p>
                          <p className="text-xs text-muted-foreground mt-1">{getErrorMessage(brpQuery.error)}</p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {/* RDW Check */}
                <div className="border border-border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <Car className="h-5 w-5 text-primary"/>
                      <h3 className="font-semibold text-foreground">RDW/BRV Controle</h3>
                    </div>
                    <Button
                      onClick={handleRDWCheck}
                      disabled={isCheckingRDW || !brpData || !!rdwData}
                      size="sm"
                    >
                      {isCheckingRDW ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                          Controleren...
                        </>
                      ) : rdwData ? (
                        <>
                          <Check className="mr-2 h-4 w-4"/>
                          Voltooid
                        </>
                      ) : (
                        "Uitvoeren"
                      )}
                    </Button>
                  </div>

                  {rdwData && (
                    <div className="bg-muted/50 rounded p-3 space-y-4 text-sm">
                      {rdwData && rdwData.length > 0 ? (
                        brpData.map((bewoner: BrpPerson, bewIndex: number) => (
                          <div key={bewIndex} className="pb-4 border-b last:border-b-0 last:pb-0">
                            <div className="font-medium text-foreground mb-2">
                              {bewoner.voornamen} {bewoner.geslachtsnaam} (BSN: {bewoner.bsn})
                            </div>
                            {rdwData.filter((b: RdwVehicle) => b.bsn === bewoner.bsn).length > 0 ? (
                              <div className="space-y-3 ml-4">
                                {rdwData.filter((b: RdwVehicle) => b.bsn === bewoner.bsn).map((voertuig: RdwVehicle, vIndex: number) => {
                                  const hidden = voertuig.diplomatiek && !isGemeentesecretaris;
                                  return (
                                    <div
                                      key={vIndex}
                                      className={`border rounded-lg p-3 ${
                                        voertuig.diplomatiek
                                          ? 'border-red-500 bg-red-500/5'
                                          : 'border-border'
                                      }`}
                                    >
                                      {voertuig.diplomatiek && (
                                        <div className="mb-3 pb-3 border-b border-red-300">
                                          <Badge className="bg-red-600 text-white">
                                            ⚠️ GEVOELIG - Diplomatiek Voertuig
                                          </Badge>
                                        </div>
                                      )}
                                      <div className="grid grid-cols-2 gap-2">
                                        <div>
                                          <span className="text-muted-foreground">Kenteken:</span>
                                          <span className={hidden ? 'ml-2 font-mono blur-sm select-none' : 'ml-2 font-mono'}>
                                            {voertuig.kenteken}
                                          </span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Voertuig:</span>
                                          <span className="ml-2">{voertuig.merk}</span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Uitstootklasse:</span>
                                          <Badge variant="outline" className="ml-2">
                                            {voertuig.klasse}
                                          </Badge>
                                        </div>
                                      </div>
                                    </div>
                                  );
                                })}
                              </div>
                            ) : (
                              <p className="text-sm text-muted-foreground ml-4">
                                Geen voertuigen geregistreerd
                              </p>
                            )}
                          </div>
                        ))
                      ) : (
                        <p className="text-sm text-muted-foreground">
                          Geen gegevens beschikbaar
                        </p>
                      )}
                    </div>
                  )}

                  {rdwQuery.isError && (
                    <div className="bg-destructive/10 border border-destructive/30 rounded p-3 text-sm">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-5 w-5 text-destructive mt-0.5 shrink-0"/>
                        <div>
                          <p className="font-medium">{getDutchErrorMessage(rdwQuery.error)}</p>
                          <p className="text-xs text-muted-foreground mt-1">{getErrorMessage(rdwQuery.error)}</p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </Card>

            {/* Handmatig Controleren */}
            {brpData && rdwData && (
              <Card className="p-6">
                <h2 className="text-xl font-bold text-foreground mb-4 flex items-center gap-2">
                  <Car className="h-5 w-5"/>
                  Voertuigen die in aanmerking komen voor een laadpaal
                </h2>

                <div className="space-y-4">
                  {(() => {
                    const elektrischeVoertuigen = rdwData.filter((v: RdwVehicle) => v.klasse === "E")
                    if (elektrischeVoertuigen.length === 0) {
                      return (
                        <div className="border border-border rounded-lg p-4 bg-muted/30">
                          <p className="text-sm text-muted-foreground">
                            Geen elektrische voertuigen (uitstootklasse E) gevonden
                          </p>
                        </div>
                      );
                    }

                    return (
                      <>
                        {elektrischeVoertuigen.map((voertuig: RdwVehicle, index: number) => {
                          const hidden = voertuig.diplomatiek && !isGemeentesecretaris;
                          return (
                          <div
                            key={index}
                            className={`border rounded-lg p-4 ${
                              voertuig.diplomatiek
                                ? 'border-red-500 bg-red-500/5'
                                : 'border-border bg-muted/30'
                            }`}
                          >
                            {voertuig.diplomatiek && (
                              <div className="mb-3 pb-3 border-b border-red-300">
                                <Badge className="bg-red-600 text-white">
                                  ⚠️ GEVOELIG - Diplomatiek Voertuig
                                </Badge>
                              </div>
                            )}
                            <div className="space-y-1">
                              <p className="text-sm font-medium text-foreground">
                                {voertuig.merk} (
                                <span className={hidden ? 'font-mono blur-sm select-none' : 'font-mono'}>
                                  {voertuig.kenteken}
                                </span>
                                )
                              </p>
                              <Badge variant="outline" className="mt-2">
                                Uitstootklasse {voertuig.klasse}
                              </Badge>
                            </div>
                          </div>
                          );
                        })}
                      </>
                    );
                  })()}
                </div>
              </Card>
            )}

            {/* Bestaande Laadpaal Controle */}
            <Card className="p-6">
              <h2 className="text-xl font-bold text-foreground mb-4 flex items-center gap-2">
                <Database className="h-5 w-5"/>
                Controle Bestaande Laadpalen
              </h2>

              <div className="border border-border rounded-lg p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <MapPin className="h-5 w-5 text-primary"/>
                    <h3 className="font-semibold text-foreground">Laadpaal Check</h3>
                  </div>
                  <Button
                    onClick={handleCheckLaadpaal}
                    disabled={isCheckingLaadpaal || !!existingLaadpaalData}
                    size="sm"
                  >
                    {isCheckingLaadpaal ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                        Controleren...
                      </>
                    ) : existingLaadpaalData ? (
                      <>
                        <Check className="mr-2 h-4 w-4"/>
                        Voltooid
                      </>
                    ) : (
                      "Uitvoeren"
                    )}
                  </Button>
                </div>

                {existingLaadpaalData && (
                  <div className={`rounded p-3 text-sm ${
                    existingLaadpaalData.bestaandeLaadpaal
                      ? 'bg-destructive/10 border border-destructive/30'
                      : 'bg-success/10 border border-success/30'
                  }`}>
                    <div className="flex items-start gap-2">
                      {existingLaadpaalData.bestaandeLaadpaal ? (
                        <XCircle className="h-5 w-5 text-destructive mt-0.5"/>
                      ) : (
                        <CheckCircle2 className="h-5 w-5 text-success mt-0.5"/>
                      )}
                      <div>
                        <p className="font-medium">
                          {existingLaadpaalData.bestaandeLaadpaal
                            ? 'Bestaande laadpaal gevonden'
                            : 'Geen bestaande laadpaal'}
                        </p>
                        <p className="text-xs text-muted-foreground mt-1">
                          {existingLaadpaalData.bestaandeLaadpaal
                            ? 'Er is al een laadpaalvergunning uitgegeven voor dit adres'
                            : 'Er is nog geen laadpaalvergunning uitgegeven voor dit adres'}
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {laadpalenQuery.isError && (
                  <div className="bg-destructive/10 border border-destructive/30 rounded p-3 text-sm">
                    <div className="flex items-start gap-2">
                      <XCircle className="h-5 w-5 text-destructive mt-0.5 shrink-0"/>
                      <div>
                        <p className="font-medium">{getDutchErrorMessage(laadpalenQuery.error)}</p>
                        <p className="text-xs text-muted-foreground mt-1">{getErrorMessage(laadpalenQuery.error)}</p>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </Card>

            {/* Vergunningbeslissing */}
            <Card className="p-6">
              <h2 className="text-xl font-bold text-foreground mb-4 flex items-center gap-2">
                <FileText className="h-5 w-5"/>
                Vergunningbeslissing
              </h2>

              <div className="space-y-4">
                {(() => {
                  if (diplomaatCheckFailed) {
                    return (
                      <div className="bg-destructive/10 border border-destructive/30 rounded p-3 text-sm">
                        <div className="flex items-start gap-2">
                          <XCircle className="h-5 w-5 text-destructive mt-0.5 shrink-0"/>
                          <div>
                            <p className="font-medium">{getDutchErrorMessage(diplomaatQuery.error)}</p>
                            <p className="text-xs text-muted-foreground mt-1">{getErrorMessage(diplomaatQuery.error)}</p>
                          </div>
                        </div>
                      </div>
                    );
                  }

                  if (diplomaatCheckPending) {
                    return (
                      <div className="border border-border rounded-lg p-4 bg-muted/30">
                        <div className="flex items-start gap-3">
                          <Loader2 className="h-5 w-5 text-muted-foreground mt-0.5 animate-spin"/>
                          <p className="text-sm text-muted-foreground">Autorisatie voor deze aanvraag wordt gecontroleerd...</p>
                        </div>
                      </div>
                    );
                  }

                  if (requiresAdditionalAuthorization) {
                    return (
                      <div className="border border-destructive/50 rounded-lg p-4 bg-destructive/5">
                        <div className="flex items-start gap-3">
                          <AlertCircle className="h-5 w-5 text-destructive mt-0.5"/>
                          <div>
                            <p className="font-semibold text-foreground mb-1">
                              Er kan niet over deze aanvraag worden besloten
                            </p>
                            <p className="text-sm text-muted-foreground">
                              De aanvraag moet worden ingediend bij een functionaris met aanvullende autorisatie
                            </p>
                          </div>
                        </div>
                      </div>
                    );
                  }

                  return (
                    <>
                      {/* Validatie checks */}
                      <div className="border border-border rounded-lg p-4 space-y-3">
                        <h3 className="font-semibold text-foreground mb-3">Validatie Status</h3>

                        {/* Check 1: Elektrisch voertuig */}
                        <div className="flex items-start gap-3">
                          {(() => {
                            // A diplomatic electric vehicle counts once the current employee is authorized to act on it.
                            const hasVisibleElectricVehicle = rdwData?.some((b: RdwVehicle) =>
                              b.klasse === "E" && (!b.diplomatiek || isGemeentesecretaris)
                            );
                            const hasUnauthorizedDiplomaticElectric = rdwData?.some((b: RdwVehicle) =>
                              b.klasse === "E" && b.diplomatiek && !isGemeentesecretaris
                            );

                            if (!rdwData) {
                              return <Clock className="h-5 w-5 text-muted-foreground mt-0.5"/>;
                            }
                            if (hasVisibleElectricVehicle) {
                              return <CheckCircle2 className="h-5 w-5 text-success mt-0.5"/>;
                            }
                            if (hasUnauthorizedDiplomaticElectric) {
                              return <AlertCircle className="h-5 w-5 text-warning mt-0.5"/>;
                            }
                            return <XCircle className="h-5 w-5 text-destructive mt-0.5"/>;
                          })()}
                          <div className="flex-1">
                            <p className="font-medium text-sm">Elektrisch voertuig</p>
                            <p className="text-xs text-muted-foreground">
                              {!rdwData
                                ? "Voer eerst RDW controle uit"
                                : (() => {
                                  const hasVisibleElectricVehicle = rdwData?.some((b: RdwVehicle) =>
                                    b.klasse === "E" && (!b.diplomatiek || isGemeentesecretaris)
                                  );
                                  const hasUnauthorizedDiplomaticElectric = rdwData?.some((b: RdwVehicle) =>
                                    b.klasse === "E" && b.diplomatiek && !isGemeentesecretaris
                                  );

                                  if (hasVisibleElectricVehicle) {
                                    return "Er is een elektrisch voertuig geregistreerd (uitstootklasse E)";
                                  }
                                  if (hasUnauthorizedDiplomaticElectric) {
                                    return "Aanvullende autorisatie vereist voor toegang tot voertuiggegevens";
                                  }
                                  return "Geen elektrisch voertuig geregistreerd";
                                })()
                              }
                            </p>
                          </div>
                        </div>

                        {/* Check 2: Geen bestaande laadpaal */}
                        <div className="flex items-start gap-3">
                          {existingLaadpaalData && !existingLaadpaalData.bestaandeLaadpaal ? (
                            <CheckCircle2 className="h-5 w-5 text-success mt-0.5"/>
                          ) : existingLaadpaalData?.bestaandeLaadpaal ? (
                            <XCircle className="h-5 w-5 text-destructive mt-0.5"/>
                          ) : (
                            <Clock className="h-5 w-5 text-muted-foreground mt-0.5"/>
                          )}
                          <div className="flex-1">
                            <p className="font-medium text-sm">Geen bestaande laadpaal</p>
                            <p className="text-xs text-muted-foreground">
                              {!existingLaadpaalData
                                ? "Voer eerst laadpaal controle uit"
                                : existingLaadpaalData.bestaandeLaadpaal
                                  ? "Er is al een laadpaal op dit adres"
                                  : "Geen bestaande laadpaal gevonden"
                              }
                            </p>
                          </div>
                        </div>
                      </div>

                      {/* Beslissing knoppen */}
                      <div className="space-y-3">
                        <Label className="text-base font-semibold">Beslissing</Label>
                        <p className="text-sm text-muted-foreground mb-3">
                          Op basis van de controles, kan deze vergunningaanvraag worden toegekend?
                        </p>
                        <div className="flex gap-3">
                          <Button
                            variant={vergunningBeslissing === "goedkeuren" ? "default" : "outline"}
                            className={`flex-1 ${vergunningBeslissing === "goedkeuren" ? "bg-success hover:bg-success/90" : ""}`}
                            onClick={() => setVergunningBeslissing("goedkeuren")}
                            disabled={
                              !rdwData ||
                              !rdwData?.some((b: RdwVehicle) =>
                                b.klasse === "E" && (!b.diplomatiek || isGemeentesecretaris)
                              ) ||
                              !existingLaadpaalData ||
                              existingLaadpaalData.bestaandeLaadpaal
                            }
                          >
                            <CheckCircle2 className="mr-2 h-4 w-4"/>
                            Goedkeuren
                          </Button>
                          <Button
                            variant={vergunningBeslissing === "afwijzen" ? "destructive" : "outline"}
                            className="flex-1"
                            onClick={() => setVergunningBeslissing("afwijzen")}
                          >
                            <XCircle className="mr-2 h-4 w-4"/>
                            Afwijzen
                          </Button>
                        </div>

                        {vergunningBeslissing && (
                          <div className={`mt-3 p-3 rounded-lg ${
                            vergunningBeslissing === "goedkeuren"
                              ? "bg-success/10 border border-success/30"
                              : "bg-destructive/10 border border-destructive/30"
                          }`}>
                            <p className="text-sm font-medium">
                              {vergunningBeslissing === "goedkeuren"
                                ? "✓ Vergunningaanvraag wordt goedgekeurd"
                                : "✗ Vergunningaanvraag wordt afgewezen"}
                            </p>
                          </div>
                        )}
                      </div>
                    </>
                  );
                })()}
              </div>
            </Card>

            {/* Communication with Zaaksysteem */}
            {application.status >= 0 && application.status <= 4 && (
              <Card className="p-6">
                <h2 className="text-xl font-bold text-foreground mb-4 flex items-center gap-2">
                  <FileText className="h-5 w-5"/>
                  Vastleggen in Zaaksysteem
                </h2>

                <div className="space-y-4">
                  <div>
                    <Label htmlFor="zaaknummer">Zaaknummer</Label>
                    <Input
                      id="zaaknummer"
                      value={zaaknummer}
                      onChange={(e) => setZaaknummer(e.target.value)}
                      className="font-mono"
                    />
                  </div>

                  <div>
                    <Label htmlFor="toelichting">Toelichting voor zaaksysteem</Label>
                    <Textarea
                      id="toelichting"
                      placeholder="Voeg een toelichting toe over de uitgevoerde controles en beslissing..."
                      value={toelichting}
                      onChange={(e) => setToelichting(e.target.value)}
                      rows={4}
                    />
                  </div>

                  {updateError && (
                    <Alert variant="destructive">
                      <AlertCircle className="h-4 w-4"/>
                      <AlertDescription>
                        {updateError}
                      </AlertDescription>
                    </Alert>
                  )}

                  {(() => {
                    // Only a confirmed denial hands off to another functionary - while the pdp2
                    // check is still pending or failed, keep the normal button but disabled,
                    // rather than presenting an unresolved check as a confirmed denial.
                    if (requiresAdditionalAuthorization && !diplomaatCheckPending && !diplomaatCheckFailed) {
                      return (
                        <Button
                          onClick={() => {
                          }}
                          className="w-full"
                        >
                          <Send className="mr-2 h-4 w-4"/>
                          Indienen aanvraag bij functionaris met aanvullende autorisatie
                        </Button>
                      );
                    }

                    return (
                      <Button
                        onClick={handleBehandelingAfronden}
                        disabled={isSendingToZaak || !vergunningBeslissing || requiresAdditionalAuthorization}
                        className="w-full"
                      >
                        {isSendingToZaak ? (
                          <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin"/>
                            Verwerken...
                          </>
                        ) : (
                          <>
                            <Send className="mr-2 h-4 w-4"/>
                            Behandeling afronden
                          </>
                        )}
                      </Button>
                    );
                  })()}
                </div>
              </Card>
            )}

            {/* Bevestigingsdialoog */}
            <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Behandeling afronden</AlertDialogTitle>
                  <AlertDialogDescription className="space-y-3">
                    <p className="font-semibold text-foreground">
                      {vergunningBeslissing === "goedkeuren"
                        ? "Laadpaalvergunning toegewezen"
                        : "Laadpaalvergunning afgewezen"}
                    </p>
                    <p>
                      Voor adres: <span
                      className="font-medium">{getFullAddress(application.postcode, application.huisnummer)}</span>
                    </p>
                    <div className="pt-2 border-t">
                      <p className="text-sm font-medium mb-2">Op basis van:</p>
                      <ul className="text-sm space-y-1 list-disc list-inside">
                        {brpData && (
                          brpData.some((b) => b.volwassen) ? (
                            <li>BRP controle: Meerderjarige bewoner gevonden en ingeschreven</li>
                          ) : (
                            <li>BRP controle: Geen meerderjarige bewoner gevonden</li>
                          )
                        )}
                        {rdwData?.some((b: RdwVehicle) => b.klasse === "E") ? (
                          <li>RDW controle: Elektrisch voertuig geregistreerd</li>
                        ) : rdwData ? (
                          <li>RDW controle: Geen elektrisch voertuig gevonden</li>
                        ) : null}
                        {existingLaadpaalData && !existingLaadpaalData.bestaandeLaadpaal ? (
                          <li>Laadpaal controle: Geen bestaande laadpaal op dit adres</li>
                        ) : existingLaadpaalData && (
                          <li>Laadpaal controle: Er bestaat al een laadpaal op dit adres</li>
                        )}
                      </ul>
                    </div>
                    <div className="pt-2 border-t">
                      <p className="text-sm font-medium mb-2">Beslissing is genomen door bevoegd ambtenaar:</p>
                      <ul className="text-sm space-y-1 list-disc list-inside">
                        <li>van de gemeente Vlierdam</li>
                        <li>
                          {isGemeentesecretaris
                            ? "werkzaam als gemeentefunctionaris met verhoogde autorisatie izv inwonende diplomaten"
                            : "werkzaam op de afdeling Openbare ruimte en energievoorziening"
                          }
                        </li>
                        <li>gediplomeerd volgens &apos;Laadpalen criterium 2025A&apos;</li>
                        <li>binnen kantoortijden</li>
                      </ul>
                    </div>
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Annuleren</AlertDialogCancel>
                  <AlertDialogAction onClick={confirmSendToZaaksysteem}>
                    Bevestigen en afronden
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>

          {/* Feedback Column - always visible */}
          <div className="lg:col-span-1">
            <Card
              className="p-6 sticky top-6 max-h-[calc(100vh-3rem)] flex flex-col rounded-none bg-transparent border-0">

              {/* Reserved space for sticky item */}
              <div className="mb-4 flex-shrink-0">
                {(() => {
                  // Als de aanvraag is verzonden, toon de samenvatting
                  if (isVerzonden && vergunningBeslissing) {
                    const statusIcon = vergunningBeslissing === "goedkeuren"
                      ? <CheckCircle2 className="h-5 w-5 text-success"/>
                      : <XCircle className="h-5 w-5 text-destructive"/>;
                    const bgColor = vergunningBeslissing === "goedkeuren"
                      ? "bg-success/10"
                      : "bg-destructive/10";
                    const borderColor = vergunningBeslissing === "goedkeuren"
                      ? "border-success/30"
                      : "border-destructive/30";

                    return (
                      <div className={`border ${borderColor} rounded-none p-4 text-sm ${bgColor}`}>
                        <div className="flex items-start gap-3">
                          {statusIcon}
                          <div className="flex-1">
                            <p className="font-semibold text-foreground">
                              {vergunningBeslissing === "goedkeuren"
                                ? "Laadpaalvergunning toegewezen"
                                : "Laadpaalvergunning afgewezen"}
                            </p>
                            <p className="text-xs text-muted-foreground mt-2">
                              Voor adres: <span
                              className="font-medium">{getFullAddress(application.postcode, application.huisnummer)}</span>
                            </p>
                            <div className="pt-2 mt-2 border-t border-current/20">
                              <p className="text-xs font-medium mb-1">Op basis van:</p>
                              <ul className="text-xs space-y-0.5 list-disc list-inside text-muted-foreground">
                                {brpData && (
                                  brpData.some((b) => b.volwassen) ? (
                                    <li>BRP controle: Meerderjarige bewoner gevonden en ingeschreven</li>
                                  ) : (
                                    <li>BRP controle: Geen meerderjarige bewoner gevonden</li>
                                  )
                                )}
                                {rdwData?.some((b: RdwVehicle) => b.klasse === "E") ? (
                                  <li>RDW controle: Elektrisch voertuig geregistreerd</li>
                                ) : rdwData ? (
                                  <li>RDW controle: Geen elektrisch voertuig gevonden</li>
                                ) : null}
                                {existingLaadpaalData && !existingLaadpaalData.bestaandeLaadpaal ? (
                                  <li>Laadpaal controle: Geen bestaande laadpaal op dit adres</li>
                                ) : existingLaadpaalData && (
                                  <li>Laadpaal controle: Er bestaat al een laadpaal op dit adres</li>
                                )}
                              </ul>
                            </div>
                            <div className="pt-2 mt-2 border-t border-current/20">
                              <p className="text-xs font-medium mb-1">Beslissing genomen door:</p>
                              <ul className="text-xs space-y-0.5 list-disc list-inside text-muted-foreground">
                                <li>Bevoegd ambtenaar van de gemeente Vlierdam</li>
                                <li>
                                  {isGemeentesecretaris
                                    ? "Werkzaam als gemeentefunctionaris met verhoogde autorisatie izv inwonende diplomaten"
                                    : "Werkzaam op de afdeling Openbare ruimte en energievoorziening"
                                  }
                                </li>
                                <li>Gediplomeerd volgens &apos;Laadpalen criterium 2025A&apos;</li>
                                <li>Binnen kantoortijden</li>
                              </ul>
                            </div>
                          </div>
                        </div>
                      </div>
                    );
                  }

                  // Anders toon de normale status
                  const hasElectricVehicle = !!rdwData?.some((b: RdwVehicle) => b.klasse === "E");
                  const noExistingLaadpaal = existingLaadpaalData && !existingLaadpaalData.bestaandeLaadpaal;
                  const allChecksCompleted = brpData && rdwData && existingLaadpaalData;
                  const canApprove = hasElectricVehicle && noExistingLaadpaal;
                  const canReject = allChecksCompleted && !canApprove;
                  const readyToDecide = allChecksCompleted;

                  // Check for mismatch between decision and check results
                  const hasMismatch = vergunningBeslissing && allChecksCompleted && (
                    (vergunningBeslissing === "goedkeuren" && !canApprove) ||
                    (vergunningBeslissing === "afwijzen" && canApprove)
                  );

                  let statusText = "Controles worden uitgevoerd...";
                  let statusIcon = <Clock className="h-5 w-5 text-muted-foreground"/>;
                  let bgColor = "bg-muted/30";
                  let borderColor = "border-border";

                  if (diplomaatCheckFailed) {
                    statusText = "Autorisatiecontrole mislukt";
                    statusIcon = <XCircle className="h-5 w-5 text-destructive"/>;
                    bgColor = "bg-destructive/10";
                    borderColor = "border-destructive/30";
                  } else if (diplomaatCheckPending) {
                    statusText = "Autorisatie wordt gecontroleerd...";
                    statusIcon = <Loader2 className="h-5 w-5 text-muted-foreground animate-spin"/>;
                  } else if (requiresAdditionalAuthorization) {
                    statusText = "Er kan niet over deze aanvraag worden besloten";
                    statusIcon = <AlertCircle className="h-5 w-5 text-destructive"/>;
                    bgColor = "bg-destructive/10";
                    borderColor = "border-destructive/30";
                  } else if (hasMismatch) {
                    statusText = "Gekozen beslissing komt niet overeen met de uitkomst van de controles";
                    statusIcon = <AlertCircle className="h-5 w-5 text-warning"/>;
                    bgColor = "bg-warning/10";
                    borderColor = "border-warning/30";
                  } else if (allChecksCompleted) {
                    if (canApprove) {
                      statusText = "Aanvraag kan worden goedgekeurd";
                      statusIcon = <CheckCircle2 className="h-5 w-5 text-success"/>;
                      bgColor = "bg-success/10";
                      borderColor = "border-success/30";
                    } else if (canReject) {
                      statusText = "Aanvraag kan worden afgewezen";
                      statusIcon = <XCircle className="h-5 w-5 text-destructive"/>;
                      bgColor = "bg-destructive/10";
                      borderColor = "border-destructive/30";
                    }
                  } else if (!brpData && !rdwData && !existingLaadpaalData) {
                    statusText = "Start met het uitvoeren van controles";
                    statusIcon = <AlertCircle className="h-5 w-5 text-warning"/>;
                    bgColor = "bg-warning/10";
                    borderColor = "border-warning/30";
                  }

                  return (
                    <div className={`border ${borderColor} rounded-none p-4 text-sm ${bgColor}`}>
                      <div className="flex items-start gap-3">
                        {statusIcon}
                        <div className="flex-1">
                          <p className="font-semibold text-foreground">{statusText}</p>
                          {diplomaatCheckFailed && (
                            <p className="text-xs text-muted-foreground mt-1">{getErrorMessage(diplomaatQuery.error)}</p>
                          )}
                          {requiresAdditionalAuthorization && !diplomaatCheckPending && !diplomaatCheckFailed && (
                            <p className="text-xs text-muted-foreground mt-1">
                              De aanvraag moet worden ingediend bij een functionaris met aanvullende autorisatie
                            </p>
                          )}
                          {!requiresAdditionalAuthorization && hasMismatch && (
                            <p className="text-xs text-muted-foreground mt-1">
                              De gekozen beslissing komt niet overeen met de controleresultaten. Weet u zeker dat u wilt
                              afwijken?
                            </p>
                          )}
                          {!requiresAdditionalAuthorization && !hasMismatch && readyToDecide && (
                            <p className="text-xs text-muted-foreground mt-1">
                              Alle informatie is beschikbaar. U kunt nu een beslissing nemen en verzenden naar het
                              zaaksysteem.
                            </p>
                          )}
                          {!requiresAdditionalAuthorization && !allChecksCompleted && (
                            <div className="mt-2">
                              <p className="text-xs font-medium mb-1">Nog uit te voeren:</p>
                              <ul className="text-xs space-y-0.5 list-disc list-inside text-muted-foreground">
                                {!brpData && <li>BRP controle uitvoeren</li>}
                                {brpData && !rdwData && <li>RDW controle uitvoeren</li>}
                                {brpData && rdwData && !existingLaadpaalData && <li>Laadpaal controle uitvoeren</li>}
                              </ul>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })()}
              </div>
            </Card>
          </div>
        </div>
      </main>

      {/* Technical Log Details Modal */}
      <Dialog open={!!selectedLogEntry} onOpenChange={() => setSelectedLogEntry(null)}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Technische Log Details</DialogTitle>
            <DialogDescription>
              {selectedLogEntry?.authzenDecision
                ? "Authorization Decision Log volgens NLGov AuthZEN standaard"
                : "Data Processing Log volgens NLGov LDV standaard"}
            </DialogDescription>
          </DialogHeader>

          {selectedLogEntry && (
            <div className="space-y-4 text-sm">
              {/* Request/Response Section */}
              <div className="space-y-2">
                <h3 className="font-semibold">Verzoek & Beslissing</h3>
                <div className="grid gap-2 p-3 bg-muted rounded-md">
                  <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                    <span className="font-medium">Timestamp:</span>
                    <span className="font-mono text-xs">{selectedLogEntry.timestamp.toISOString()}</span>
                  </div>
                  <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                    <span className="font-medium">Type:</span>
                    <span>{selectedLogEntry.action}</span>
                  </div>
                  <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                    <span className="font-medium">Resultaat:</span>
                    <Badge variant={
                      selectedLogEntry.result === "success" ? "default" :
                        selectedLogEntry.result === "error" ? "destructive" : "secondary"
                    } className="w-fit">
                      {selectedLogEntry.result === "success" ? "Toegestaan" :
                        selectedLogEntry.result === "error" ? "Geweigerd" : "Informatief"}
                    </Badge>
                  </div>
                </div>
              </div>

              {/* System Architecture Section */}
              <div className="space-y-2">
                <h3 className="font-semibold">Systeem Architectuur</h3>
                <div className="grid gap-2 p-3 bg-muted rounded-md">
                  <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                    <span className="font-medium">
                      {selectedLogEntry.authzenDecision ? "Toegangssysteem:" : "Bronsysteem:"}
                    </span>
                    <Badge variant="default" className="w-fit">{selectedLogEntry.system}</Badge>
                  </div>
                  <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                    <span className="font-medium">Bronhouder:</span>
                    <span className="text-sm">
                      {selectedLogEntry.system === "BRP" && "RvIG (Rijksdienst voor Identiteitsgegevens)"}
                      {selectedLogEntry.system === "BRV" && "RDW (Rijksdienst voor het Wegverkeer)"}
                      {selectedLogEntry.system === "FDS" && "Kadaster"}
                      {selectedLogEntry.system === "GLP" && "Gemeente Vlierdam"}
                      {selectedLogEntry.system === "Zaaksysteem" && "Gemeente Vlierdam"}
                      {selectedLogEntry.system === "FTV" && "Logius"}
                      {selectedLogEntry.system === "FSC" && "Logius"}
                      {selectedLogEntry.system === "LDV" && "Gemeente Vlierdam"}
                      {selectedLogEntry.system === "LTV" && "Logius"}
                      {selectedLogEntry.system === "Ambtenaar" && "Gemeente Vlierdam"}
                    </span>
                  </div>
                  {selectedLogEntry.connectionType && (
                    <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                      <span className="font-medium">Verbinding:</span>
                      <Badge variant={selectedLogEntry.connectionType === "FSC" ? "default" : "secondary"}
                             className="w-fit">
                        {selectedLogEntry.connectionType === "FSC"
                          ? "FSC (Federatief Service Connectiviteit)"
                          : "Intern Gemeentelijk"}
                      </Badge>
                    </div>
                  )}
                  {selectedLogEntry.accessControl && (
                    <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                      <span className="font-medium">Toegangscontrole:</span>
                      <Badge variant="default" className="w-fit">{selectedLogEntry.accessControl} (Federatief
                        ToegangsVerlenings Register)</Badge>
                    </div>
                  )}
                  {selectedLogEntry.logging && (
                    <div className="grid grid-cols-[140px_1fr] gap-2 items-center">
                      <span className="font-medium">Logging Type:</span>
                      <Badge variant="secondary" className="w-fit">
                        {selectedLogEntry.logging === "LDV"
                          ? "LDV (Logboek DataVerwerkingen)"
                          : "LTV (Logboek ToegangsVerlening)"}
                      </Badge>
                    </div>
                  )}
                </div>
              </div>

              {/* Trace Context Section */}
              <div className="space-y-2">
                <h3 className="font-semibold">Trace Context (W3C)</h3>
                <div className="grid gap-2 p-3 bg-muted rounded-md font-mono text-xs">
                  <div className="grid grid-cols-[140px_1fr] gap-2">
                    <span className="font-sans font-medium">trace_id:</span>
                    <span className="break-all">{selectedLogEntry.id.split('-').slice(0, 4).join('-')}</span>
                  </div>
                  <div className="grid grid-cols-[140px_1fr] gap-2">
                    <span className="font-sans font-medium">span_id:</span>
                    <span>{selectedLogEntry.id.split('-')[4]}</span>
                  </div>
                  {application?.id && (
                    <div className="grid grid-cols-[140px_1fr] gap-2">
                      <span className="font-sans font-medium">transaction_id:</span>
                      <span>{application.id}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* AuthZEN Request Section - Only for FTV logs */}
              {selectedLogEntry.authzenDecision && (
                <>
                  <div className="space-y-2">
                    <h3 className="font-semibold">AuthZEN Request</h3>
                    <div className="p-3 bg-muted rounded-md">
                      <pre className="text-xs overflow-x-auto whitespace-pre-wrap">
{JSON.stringify({
  subject: {
    type: "user",
    id: currentUser,
    attributes: {
      role: isGemeentesecretaris ? "gemeentesecretaris" : "ambtenaar",
      department: "Openbare ruimte en energievoorziening"
    }
  },
  action: {
    name: selectedLogEntry.action
  },
  resource: {
    type: selectedLogEntry.system,
    id: application?.id,
    attributes: {
      address: `${application?.postcode} ${application?.huisnummer}`
    }
  },
  context: {
    time: selectedLogEntry.timestamp.toISOString()
  }
}, null, 2)}
                      </pre>
                    </div>
                  </div>

                  {/* AuthZEN Response Section - Only for FTV logs */}
                  <div className="space-y-2">
                    <h3 className="font-semibold">AuthZEN Response</h3>
                    <div className="p-3 bg-muted rounded-md">
                      <pre className="text-xs overflow-x-auto whitespace-pre-wrap">
{JSON.stringify({
  decision: selectedLogEntry.authzenDecision,
  context: {
    id: selectedLogEntry.id,
    reason_admin: {
      nl: selectedLogEntry.authzenReason || selectedLogEntry.details
    }
  }
}, null, 2)}
                      </pre>
                    </div>
                  </div>
                </>
              )}

              {/* Data Response Section - Only for bronsysteem logs */}
              {selectedLogEntry.dataResponse && (
                <div className="space-y-2">
                  <h3 className="font-semibold">Data Response</h3>
                  <div className="p-3 bg-muted rounded-md">
                    <pre className="text-xs overflow-x-auto whitespace-pre-wrap">
{JSON.stringify({
  records_found: selectedLogEntry.dataResponse.recordsFound,
  summary: selectedLogEntry.dataResponse.summary,
  timestamp: selectedLogEntry.timestamp.toISOString(),
  trace_id: selectedLogEntry.id.split('-').slice(0, 4).join('-')
}, null, 2)}
                    </pre>
                  </div>
                </div>
              )}

              {/* Public Details Section */}
              {selectedLogEntry.isPublic && (
                <div className="space-y-2">
                  <h3 className="font-semibold">Openbare Feedback</h3>
                  <Alert>
                    <Info className="h-4 w-4"/>
                    <AlertDescription>{selectedLogEntry.details}</AlertDescription>
                  </Alert>
                </div>
              )}

              {/* Policy & Information Sources */}
              <div className="space-y-2">
                <h3 className="font-semibold">Policy & Information Sources</h3>
                <div className="grid gap-2 p-3 bg-muted rounded-md">
                  <div>
                    <span className="font-medium">Policies:</span>
                    <div className="mt-1 text-xs font-mono">
                      {JSON.stringify({
                        "laadpaal-policy-repo": {
                          version: "2.1.0",
                          filter: "laadpalen-criteria-2025A"
                        }
                      }, null, 2)}
                    </div>
                  </div>
                  {selectedLogEntry.connectionType === "FSC" && (
                    <div className="mt-2">
                      <span className="font-medium">Information Sources:</span>
                      <div className="mt-1 text-xs font-mono">
                        {JSON.stringify({
                          [selectedLogEntry.system.toLowerCase()]: {
                            endpoint: `https://fsc.overheid.nl/${selectedLogEntry.system.toLowerCase()}/v1`,
                            span_id: selectedLogEntry.id.split('-')[4]
                          }
                        }, null, 2)}
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Engine Configuration - Only for FTV logs */}
              {selectedLogEntry.authzenDecision && (
                <div className="space-y-2">
                  <h3 className="font-semibold">Engine Configuratie (PDP)</h3>
                  <div className="p-3 bg-muted rounded-md font-mono text-xs">
                    {JSON.stringify({
                      type: "OPA",
                      version: "0.68.0",
                      hostname: "pdp-laadpalen-001.vlierdam.nl",
                      config: {
                        decision_log_enabled: true,
                        bundle_download_interval: "30s"
                      }
                    }, null, 2)}
                  </div>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default AanvraagDetail;