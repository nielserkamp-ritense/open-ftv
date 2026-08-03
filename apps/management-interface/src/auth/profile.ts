// Profile claims live in the ID token (scope `profile`), not the access token — the
// Keycloak realm maps roles into the access token only. See ADR 0002.
export interface ProfileClaims {
  name?: string;
  preferred_username?: string;
  email?: string;
}

export const UNKNOWN_USER = 'Onbekende gebruiker';

// displayName picks the most human-readable name the IdP handed us, falling back down
// the chain until something is present.
export const displayName = (profile?: ProfileClaims): string => {
  const candidates = [profile?.name, profile?.preferred_username, profile?.email];
  return candidates.map((c) => c?.trim()).find((c) => c) ?? UNKNOWN_USER;
};

// initials renders a name as at most two uppercase letters for the header avatar:
// the first letter of the first and last word, or the first two letters of a single word.
export const initials = (name: string): string => {
  const words = name.trim().split(/\s+/).filter(Boolean);
  const first = words[0];
  const last = words[words.length - 1];
  if (first === undefined || last === undefined) return '?';
  if (words.length === 1) return first.slice(0, 2).toUpperCase();
  return (first.slice(0, 1) + last.slice(0, 1)).toUpperCase();
};
