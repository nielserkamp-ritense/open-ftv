import type { AuthlogEntry } from './authlog';

export type Decision = 'permitted' | 'denied' | 'error';

export const evaluationFailed = 'Evaluatie mislukt';

interface DecisionContext {
  reason?: unknown;
  reason_user?: Record<string, string>;
  reason_admin?: Record<string, string>;
  error?: { message?: unknown };
}

interface EvaluationDecision {
  decision?: boolean;
  context?: DecisionContext;
}

// The OpenAPI types define `request` and `response` as Record<string, never>, so we need safe casts here.
interface LoggedResponse extends EvaluationDecision {
  evaluations?: EvaluationDecision[];
}

interface LoggedRequest {
  options?: { evaluations_semantic?: string };
}

const responseOf = (entry: AuthlogEntry) => entry.response as unknown as LoggedResponse | undefined;

// An evaluation in a batch that failed in the engine is denied, with the error in its context.
const failed = (e: EvaluationDecision) => e.context?.error !== undefined;

// A batch logs one entry for all its evaluations, so its decision follows the batch's semantics: with
// permit_on_first_permit one permit is enough, otherwise every evaluation must be permitted.
const batchPermitted = (entry: AuthlogEntry, evaluations: EvaluationDecision[]) => {
  const semantic = (entry.request as unknown as LoggedRequest | undefined)?.options?.evaluations_semantic;

  return semantic?.toLowerCase() === 'permit_on_first_permit'
    ? evaluations.some((e) => e.decision === true)
    : evaluations.every((e) => e.decision === true);
};

// entryDecision returns undefined for entries without a decision, like searches.
export const entryDecision = (entry?: AuthlogEntry): Decision | undefined => {
  if (!entry) return undefined;
  if (entry.status === 'Error') return 'error';

  const response = responseOf(entry);
  if (typeof response?.decision === 'boolean') {
    return response.decision ? 'permitted' : 'denied';
  }

  const evaluations = response?.evaluations;
  if (!evaluations?.length) return undefined;

  if (batchPermitted(entry, evaluations)) return 'permitted';
  return evaluations.some(failed) ? 'error' : 'denied';
};

const contextReason = (context?: DecisionContext): string | undefined => {
  if (typeof context?.reason === 'string' && context.reason !== '') {
    return context.reason;
  }
  if (typeof context?.error?.message === 'string' && context.error.message !== '') {
    return context.error.message;
  }
  // Decisions logged before AuthZEN 1.0 carry the reason in reason_user or reason_admin. The PDP
  // filled it with "ok" or "not authorized" when the policy gave no reason; those are not reasons.
  const reasons = context?.reason_user ?? context?.reason_admin;
  const legacy = reasons ? Object.values(reasons)[0] : undefined;
  return legacy === 'ok' || legacy === 'not authorized' ? undefined : legacy;
};

// entryReason returns, for a batch, the error of its first failed evaluation, or else the reason of its
// first denied evaluation.
export const entryReason = (entry?: AuthlogEntry): string | undefined => {
  if (!entry) return undefined;
  if (entry.status === 'Error') return evaluationFailed;

  const response = responseOf(entry);
  const evaluations = response?.evaluations;
  if (evaluations) {
    const evaluation = entryDecision(entry) === 'error' ? evaluations.find(failed) : evaluations.find((e) => e.decision === false);
    return contextReason(evaluation?.context);
  }

  return contextReason(response?.context);
};
