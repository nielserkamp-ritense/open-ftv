import {useState} from "react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {AlertCircle, CheckCircle2, XCircle} from "lucide-react";
import {z} from "zod";
import {evaluateLaadpaalIndienen, newTraceId} from "@/lib/api";

const applicationSchema = z.object({
  postcode: z.string()
    .trim()
    .regex(/^[1-9][0-9]{3}\s?[A-Z]{2}$/i, {message: "Ongeldige postcode (bijv. 1234AB)"})
    .transform(val => val.toUpperCase().replace(/\s/g, "")),
  huisnummer: z.number()
    .min(1, {message: "Huisnummer mag niet lager zijn dan 1"}),
});

interface AuditLogEntry {
  action: string;
  system: string;
  result: "success" | "error" | "info";
  details: string;
  isPublic?: boolean;
  traceId?: string;
}

interface ApplicationFormProps {
  onSubmit: (data: { postcode: string; huisnummer: number }) => void;
  onAuditLog?: (entry: AuditLogEntry) => void;
  currentUser: string;
  gebruikers: Array<{ naam: string; afdeling: string; "laadpaal-opleiding"?: string }>;
}

export const ApplicationForm = ({onSubmit, onAuditLog, currentUser, gebruikers}: ApplicationFormProps) => {
  const [postcode, setPostcode] = useState("");
  const [huisnummer, setHuisnummer] = useState(1);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [publicFeedback, setPublicFeedback] = useState<{ type: "success" | "error", message: string } | null>(null);
  const [isAuthorizing, setIsAuthorizing] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});

    try {
      const validated = applicationSchema.parse({
        postcode,
        huisnummer
      });

      const gebruiker = gebruikers.find(g => g.naam === currentUser);
      const traceId = newTraceId();
      setIsAuthorizing(true);
      let beslissing;
      try {
        beslissing = await evaluateLaadpaalIndienen({
          gebruiker: currentUser,
          afdeling: gebruiker?.afdeling,
          laadpaalOpleiding: gebruiker?.["laadpaal-opleiding"],
          postcode: validated.postcode,
          huisnummer: validated.huisnummer,
          traceId,
        });
      } catch (error) {
        // The PDP logs a failed check as well, so its trace ID leads to the error in the Logboek.
        const feedback = {
          type: "error" as const,
          message: `Autorisatiecontrole mislukt: ${error instanceof Error ? error.message : "onbekende fout"}`,
        };
        setPublicFeedback(feedback);
        if (onAuditLog) {
          onAuditLog({
            action: "Autorisatiecontrole",
            system: "FTV",
            result: "error",
            details: feedback.message,
            isPublic: true,
            traceId,
          });
        }
        return;
      } finally {
        setIsAuthorizing(false);
      }

      if (!beslissing.decision) {
        const authorizationReason = beslissing.context?.reason ?? "onbekende reden";

        const feedback = {
          type: "error" as const,
          message: `${currentUser} niet geautoriseerd vanwege ${authorizationReason}`
        };
        setPublicFeedback(feedback);

        if (onAuditLog) {
          onAuditLog({
            action: "Autorisatiecontrole",
            system: "FTV",
            result: "error",
            details: feedback.message,
            isPublic: true,
            traceId,
          });
        }
        return;
      }

      onSubmit({
        postcode: validated.postcode,
        huisnummer: validated.huisnummer
      });

      // Reset form
      setPostcode("");
      setHuisnummer(1);
      setPublicFeedback(null);
    } catch (error) {
      if (error instanceof z.ZodError) {
        const fieldErrors: Record<string, string> = {};
        error.errors.forEach((err) => {
          if (err.path[0]) {
            fieldErrors[err.path[0].toString()] = err.message;
          }
        });
        setErrors(fieldErrors);
      } else {
        setPublicFeedback({
          type: "error",
          message: `Autorisatiecontrole mislukt: ${error instanceof Error ? error.message : "onbekende fout"}`,
        });
      }
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {publicFeedback && (
        <div
          role="alert"
          className={`rounded-lg p-4 border-l-4 ${
            publicFeedback.type === "success"
              ? "bg-success/10 border-success"
              : "bg-destructive/10 border-destructive"
          }`}
        >
          <div className="flex items-start gap-3">
            {publicFeedback.type === "success" ? (
              <CheckCircle2 className="h-5 w-5 text-success flex-shrink-0 mt-0.5"/>
            ) : (
              <XCircle className="h-5 w-5 text-destructive flex-shrink-0 mt-0.5"/>
            )}
            <div className="flex-1">
              <p className="text-sm font-medium text-foreground">
                {publicFeedback.message}
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="postcode" className="text-base font-semibold">
          Postcode <span className="text-destructive" aria-hidden="true">*</span>
        </Label>
        <Input
          id="postcode"
          value={postcode}
          onChange={(e) => setPostcode(e.target.value)}
          placeholder="1234AB"
          className={`h-12 ${errors.postcode ? "border-destructive" : ""}`}
          maxLength={7}
          aria-required="true"
          aria-invalid={!!errors.postcode}
          aria-describedby={errors.postcode ? "postcode-error" : undefined}
        />
        {errors.postcode && (
          <p id="postcode-error" className="text-sm text-destructive flex items-center gap-1">
            <AlertCircle className="h-4 w-4"/>
            {errors.postcode}
          </p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="huisnummer" className="text-base font-semibold">
          Huisnummer <span className="text-destructive" aria-hidden="true">*</span>
        </Label>
        <Input
          id="huisnummer"
          value={huisnummer}
          type="number"
          onChange={(e) => setHuisnummer(parseInt(e.target.value))}
          placeholder="123"
          className={`h-12 ${errors.huisnummer ? "border-destructive" : ""}`}
          maxLength={10}
          aria-required="true"
          aria-invalid={!!errors.huisnummer}
          aria-describedby={errors.huisnummer ? "huisnummer-error" : undefined}
        />
        {errors.huisnummer && (
          <p id="huisnummer-error" className="text-sm text-destructive flex items-center gap-1">
            <AlertCircle className="h-4 w-4"/>
            {errors.huisnummer}
          </p>
        )}
      </div>

      <div className="pt-4">
        <Button
          type="submit"
          className="w-full h-12 text-base font-semibold"
          size="lg"
          disabled={isAuthorizing}
        >
          <CheckCircle2 className="mr-2 h-5 w-5"/>
          {isAuthorizing ? "Autorisatie controleren..." : "Aanvraag Indienen"}
        </Button>
      </div>

      <p className="text-sm text-muted-foreground text-center">
        Door deze aanvraag in te dienen verklaart u dat de verstrekte gegevens juist zijn
      </p>
    </form>
  );
};
