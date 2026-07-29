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

// Capability rules mirror the Cedar policies enforced by the manager (the PDP is the
// authoritative gate; these only drive UI affordances / defense in depth).
export const canWrite = (roles: string[]): boolean =>
  roles.includes('admin') || roles.includes('author');

export const isAdmin = (roles: string[]): boolean => roles.includes('admin');

export const canPublish = (roles: string[]): boolean => isAdmin(roles);

export const isReadOnly = (roles: string[]): boolean => !canWrite(roles);
