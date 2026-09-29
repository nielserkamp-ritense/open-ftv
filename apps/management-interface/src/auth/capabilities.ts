import { jwtDecode } from 'jwt-decode';

// rolesFromToken extracts the (top-level) `roles` claim from a Keycloak access token.
// Returns [] for an absent or malformed token.
export const rolesFromToken = (token?: string): string[] => {
  if (!token) return [];
  try {
    const claims = jwtDecode<{ roles?: string[] }>(token);
    return Array.isArray(claims.roles) ? claims.roles : [];
  } catch {
    return [];
  }
};

// Capability rules mirror the SCRUM-16 role matrix the manager's seed policies enforce (the
// PDP is the authoritative gate; these only drive UI affordances). The author (functioneel
// beheerder) edits and publishes rules and PIP data; the admin (systeembeheerder) manages
// beslispunten and settings. Replaced by PDP evaluations in SCRUM-98 / SCRUM-16.
export const canWrite = (roles: string[]): boolean => roles.includes('author');

export const isAdmin = (roles: string[]): boolean => roles.includes('admin');

export const canPublish = (roles: string[]): boolean => roles.includes('author');

export const isReadOnly = (roles: string[]): boolean => !canWrite(roles);
