// Centralized environment variable access with runtime support
// Reads from window.__ENV__ (populated at container start) and falls back to import.meta.env

// Allow TypeScript to know about window.__ENV__
declare global {
  interface Window {
    __ENV__?: Record<string, string | undefined>;
  }
}

type EnvRecord = Record<string, string | undefined>;

const w: (Window & typeof globalThis) | undefined =
  typeof window !== 'undefined' ? window : undefined;

const runtimeEnv: EnvRecord = (w?.__ENV__ ?? {}) as EnvRecord;

let buildEnv: EnvRecord = {};
try {
  // import.meta.env exists in Vite context
  const maybeEnv = (import.meta as unknown as { env?: EnvRecord }).env;
  buildEnv = maybeEnv ?? {};
} catch {
  buildEnv = {};
}

const isPlaceholder = (s: string): boolean => {
  const trimmed = s.trim();
  return /^%[A-Z0-9_]+%$/.test(trimmed) || /^\$\{?[A-Z0-9_]+\}?$/.test(trimmed);
};

export const getEnvVar = (key: string, fallback?: string): string => {
  const raw = (runtimeEnv[key] ?? buildEnv[key]);
  const val = typeof raw === 'string' && raw.trim().length > 0 && !isPlaceholder(raw) ? raw : undefined;
  return val ?? (fallback ?? '');
};

export const VITE_API_BASE_URL = getEnvVar('VITE_API_BASE_URL', 'http://localhost:8080');
export const VITE_VLIERDAM_OUTWAY_URL = getEnvVar('VITE_VLIERDAM_OUTWAY_URL', 'http://localhost:8080');
export const VITE_BRP_API_KEY = getEnvVar('VITE_BRP_API_KEY', '');
export const VITE_DOELBINDING = getEnvVar('VITE_DOELBINDING', 'laadpaal-aanvraag');
export const VITE_DOELBINDING_BRP = getEnvVar('VITE_DOELBINDING_BRP', 'identificatie-en-ingezetenschap');
export const VITE_DOELBINDING_BRV = getEnvVar('VITE_DOELBINDING_BRV', 'laadinfra-verificatie');
export const VITE_BEHEER_URL = getEnvVar('VITE_BEHEER_URL', 'http://localhost:5173');
export const VITE_RVIG_BEHEER_URL = getEnvVar('VITE_RVIG_BEHEER_URL', 'http://localhost:5174');
export const VITE_RDW_BEHEER_URL = getEnvVar('VITE_RDW_BEHEER_URL', 'http://localhost:5175');

export const VITE_VLIERDAM_PDP1_URL = getEnvVar('VITE_VLIERDAM_PDP1_URL', '');
export const VITE_VLIERDAM_PDP2_URL = getEnvVar('VITE_VLIERDAM_PDP2_URL', '');
