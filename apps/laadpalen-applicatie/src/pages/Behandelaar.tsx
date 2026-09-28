import {useState} from "react";
import {useNavigate} from "react-router-dom";
import {Header} from "@/components/Header";
import {Card} from "@/components/ui/card";
import {Button} from "@/components/ui/button";
import {Badge} from "@/components/ui/badge";
import {Input} from "@/components/ui/input";
import {
  CheckCircle2,
  XCircle,
  Clock,
  Search,
  Filter,
  FileText,
  AlertTriangle,
  Download,
  Plus,
  Loader2,
  Settings2,
} from "lucide-react";
import {format} from "date-fns";
import {nl} from "date-fns/locale";
import {generateVergunningPDF} from "@/utils/pdfGenerator";
import {useAanvragen} from "@/lib/api/aanvragen";
import type {Aanvraag} from "@/lib/api/types";
import {useCurrentUser} from "@/contexts/CurrentUserContext";

const Behandelaar = () => {
  const navigate = useNavigate();
  const {data: aanvragenData, isLoading, error} = useAanvragen();
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<"all" | Aanvraag["status"]>("all");
  const {currentUser, setCurrentUser} = useCurrentUser();
  const [showAuditLog, setShowAuditLog] = useState(true);

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

  const getStatusCounts = () => {
    if (!aanvragenData) return {pending: 0, in_behandeling: 0, approved: 0, rejected: 0};

    return {
      pending: aanvragenData.filter((a) => a.status === 1).length,
      in_behandeling: aanvragenData.filter((a) => a.status === 2 || a.status === 3).length,
      approved: aanvragenData.filter((a) => a.status === 4).length,
      rejected: aanvragenData.filter((a) => a.status === 8 || a.status === 9).length,
    };
  };

  const filteredApplications = aanvragenData?.filter((app) => {
    const matchesSearch =
      app.postcode?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      app.huisnummer?.toString().includes(searchTerm);
    const matchesStatus = statusFilter === "all" || app.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const handleDownloadPDF = (application: Aanvraag) => {
    if (application.status === 4) {
      generateVergunningPDF(application);
    }
  };

  const counts = getStatusCounts();

  return (
    <div className="min-h-screen bg-background">
      <Header
        currentUser={currentUser}
        onUserChange={setCurrentUser}
        showAuditLog={showAuditLog}
        onToggleAuditLog={() => setShowAuditLog(!showAuditLog)}
      />

      <main className="container mx-auto px-4 py-8 max-w-7xl">
        <div className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-bold text-foreground mb-2">
              Admin Panel Laadpaalvergunningen
            </h1>
            <p className="text-lg text-muted-foreground">
              Beheer aanvragen, maak nieuwe vergunningen aan en download documenten
            </p>
          </div>

          <div className="flex flex-col gap-2">
            <Button
              onClick={() => navigate("/behandelaar/nieuw")}
              size="lg"
              className="gap-2"
            >
              <Plus className="h-5 w-5"/>
              Nieuwe Aanvraag
            </Button>
            <Button
              onClick={() => navigate("/behandelaar/configuratie-status")}
              size="lg"
              variant="outline"
              className="gap-2"
            >
              <Settings2 className="h-5 w-5"/>
              Config &amp; Status Bronnen
            </Button>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary"/>
            <span className="ml-3 text-lg text-muted-foreground">Aanvragen laden...</span>
          </div>
        )}

        {/* Error State */}
        {error && (
          <Card className="p-8 border-destructive">
            <div className="flex items-center gap-4">
              <AlertTriangle className="h-8 w-8 text-destructive"/>
              <div>
                <h3 className="text-lg font-semibold text-foreground mb-1">Fout bij laden van
                  aanvragen</h3>
                <p className="text-sm text-muted-foreground">
                  Er is een fout opgetreden bij het ophalen van de aanvragen. Probeer het later
                  opnieuw.
                </p>
              </div>
            </div>
          </Card>
        )}

        {/* Content - only show when not loading and no error */}
        {!isLoading && !error && (
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <div className="lg:col-span-3 space-y-6">

              {/* Statistics Cards */}
              <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                <Card className="p-6 border-l-4 border-l-warning">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Te behandelen</p>
                      <p className="text-3xl font-bold text-foreground">{counts.pending}</p>
                    </div>
                    <Clock className="h-12 w-12 text-warning"/>
                  </div>
                </Card>

                <Card className="p-6 border-l-4 border-l-blue-500">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">In behandeling</p>
                      <p className="text-3xl font-bold text-foreground">{counts.in_behandeling}</p>
                    </div>
                    <Clock className="h-12 w-12 text-blue-500"/>
                  </div>
                </Card>

                <Card className="p-6 border-l-4 border-l-success">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Toegewezen</p>
                      <p className="text-3xl font-bold text-foreground">{counts.approved}</p>
                    </div>
                    <CheckCircle2 className="h-12 w-12 text-success"/>
                  </div>
                </Card>

                <Card className="p-6 border-l-4 border-l-destructive">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm text-muted-foreground mb-1">Afgewezen</p>
                      <p className="text-3xl font-bold text-foreground">{counts.rejected}</p>
                    </div>
                    <XCircle className="h-12 w-12 text-destructive"/>
                  </div>
                </Card>
              </div>

              {/* Applications Table */}
              <Card className="overflow-hidden">
                <div className="p-6 border-b border-border space-y-4">
                  <h2 className="text-2xl font-bold text-foreground">Aanvragen</h2>

                  {/* Search Bar */}
                  <div className="relative">
                    <Search
                      className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                    <Input
                      id="search"
                      placeholder="Zoek op postcode of huisnummer..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10"
                    />
                  </div>

                  {/* Status Filters */}
                  <div className="flex gap-2 flex-wrap">
                    <Button
                      variant={statusFilter === "all" ? "default" : "outline"}
                      onClick={() => setStatusFilter("all")}
                      className="flex items-center gap-2"
                    >
                      <Filter className="h-4 w-4"/>
                      Alle ({aanvragenData?.length ?? 0})
                    </Button>
                    <Button
                      variant={statusFilter === 1 ? "default" : "outline"}
                      onClick={() => setStatusFilter(1)}
                    >
                      Te behandelen ({counts.pending})
                    </Button>
                    <Button
                      variant={statusFilter == 2 || statusFilter === 3 ? "default" : "outline"}
                      onClick={() => setStatusFilter(3)}
                    >
                      In behandeling ({counts.in_behandeling})
                    </Button>
                    <Button
                      variant={statusFilter === 4 ? "default" : "outline"}
                      onClick={() => setStatusFilter(4)}
                    >
                      Toegewezen ({counts.approved})
                    </Button>
                    <Button
                      variant={statusFilter === 8 || statusFilter === 9 ? "default" : "outline"}
                      onClick={() => setStatusFilter(9)}
                    >
                      Afgewezen ({counts.rejected})
                    </Button>
                  </div>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-muted/50 border-b border-border">
                    <tr>
                      <th className="px-6 py-4 text-left text-sm font-semibold text-foreground">
                        Adres
                      </th>
                      <th className="px-6 py-4 text-left text-sm font-semibold text-foreground">
                        Datum
                      </th>
                      <th className="px-6 py-4 text-left text-sm font-semibold text-foreground">
                        Behandelaar
                      </th>
                      <th className="px-6 py-4 text-left text-sm font-semibold text-foreground">
                        Status
                      </th>
                      <th className="px-6 py-4 text-left text-sm font-semibold text-foreground">
                        Acties
                      </th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                    {(filteredApplications?.length ?? 0) === 0 ? (
                      <tr>
                        <td colSpan={5} className="px-6 py-12 text-center">
                          <FileText className="h-12 w-12 mx-auto text-muted-foreground mb-3"/>
                          <p className="text-muted-foreground">Geen aanvragen gevonden</p>
                        </td>
                      </tr>
                    ) : (
                      filteredApplications.map((app) => (
                        <tr
                          key={app.id}
                          className="hover:bg-muted/30 transition-colors cursor-pointer"
                          onClick={() => navigate(`/behandelaar/aanvraag/${app.id}`)}
                        >
                          <td className="px-6 py-4">
                            <div className="font-medium text-foreground">
                              {app.postcode} {app.huisnummer}
                            </div>
                            <div className="text-sm text-muted-foreground font-mono">
                              ID: {app.id}
                            </div>
                          </td>
                          <td className="px-6 py-4 text-muted-foreground">
                            {format(app["status-datum"], "d MMM yyyy", {locale: nl})}
                          </td>
                          <td className="px-6 py-4">
                            {app.naam ? (
                              <span
                                className="text-foreground font-medium">{app.naam}</span>
                            ) : (
                              <span className="text-muted-foreground text-sm">Niet toegewezen</span>
                            )}
                          </td>
                          <td className="px-6 py-4">{getStatusBadge(app.status)}</td>
                          <td className="px-6 py-4">
                            {app.status === 4 ? (
                              <Button
                                size="sm"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleDownloadPDF(app);
                                }}
                              >
                                <Download className="h-4 w-4 mr-1"/>
                                PDF
                              </Button>
                            ) : (
                              <span className="text-sm text-muted-foreground">
                                {"Geen toelichting"}
                              </span>
                            )}
                          </td>
                        </tr>
                      ))
                    )}
                    </tbody>
                  </table>
                </div>
              </Card>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default Behandelaar;
