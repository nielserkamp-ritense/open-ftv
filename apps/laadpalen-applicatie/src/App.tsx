import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider, QueryCache, MutationCache } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { toast } from "sonner";
import Index from "./pages/Index";
import Behandelaar from "./pages/Behandelaar";
import NieuweAanvraag from "./pages/NieuweAanvraag";
import AanvraagDetail from "./pages/AanvraagDetail";
import ConfiguratieStatus from "./pages/ConfiguratieStatus";
import NotFound from "./pages/NotFound";
import { CurrentUserProvider } from "./contexts/CurrentUserContext";
import { getDutchErrorMessage, getErrorMessage, isApiError } from "@/lib/api/errors";

// How long an error toast stays up before auto-dismissing.
const ERROR_TOAST_DURATION_MS = 15000;

function notifyApiError(error: unknown) {
  const status = isApiError(error) ? error.status : undefined;
  const source = isApiError(error) ? error.source : undefined;

  const parts = [getErrorMessage(error)];
  if (status) parts.push(`status ${status}`);
  if (source) parts.push(`bron: ${source}`);
  const description = parts.join(" — ");

  // A stable id per distinct error condition, so sonner's own dedup replaces a still-open
  // toast for the same condition instead of stacking another - never something this app
  // decides itself by inspecting what's currently shown.
  toast.error(getDutchErrorMessage(error), {
    id: `api-error:${description}`,
    description,
    duration: ERROR_TOAST_DURATION_MS,
  });
}

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: notifyApiError,
  }),
  mutationCache: new MutationCache({
    onError: notifyApiError,
  }),
  defaultOptions: {
    queries: {
      // A 401/403 means the PDP denied the request - retrying the same call won't change that decision.
      retry: (failureCount, error) =>
        isApiError(error) && (error.status === 401 || error.status === 403) ? false : failureCount < 3,
    },
  },
});

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <CurrentUserProvider>
        <Sonner />
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Index />} />
            <Route path="/behandelaar" element={<Behandelaar />} />
            <Route path="/behandelaar/nieuw" element={<NieuweAanvraag />} />
            <Route path="/behandelaar/aanvraag/:id" element={<AanvraagDetail />} />
            <Route path="/behandelaar/configuratie-status" element={<ConfiguratieStatus />} />
            {/* ADD ALL CUSTOM ROUTES ABOVE THE CATCH-ALL "*" ROUTE */}
            <Route path="*" element={<NotFound />} />
          </Routes>
        </BrowserRouter>
      </CurrentUserProvider>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
