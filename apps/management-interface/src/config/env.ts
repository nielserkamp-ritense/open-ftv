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

export const PAP_BASE_URL = getEnvVar('VITE_PAP_BASE_URL', 'http://localhost:8080');
export const PIP_BASE_URL = getEnvVar('VITE_PIP_BASE_URL', 'http://localhost:8080');
export const OIDC_AUTHORITY = getEnvVar('VITE_OIDC_AUTHORITY', 'http://localhost:8088/realms/openftv');
export const OIDC_CLIENT_ID = getEnvVar('VITE_OIDC_CLIENT_ID', 'openftv');
