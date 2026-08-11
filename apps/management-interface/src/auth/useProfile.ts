import { useAuth } from 'react-oidc-context';
import { rolesFromToken } from './capabilities';
import { displayName, initials } from './profile';

export interface UserProfile {
  displayName: string;
  initials: string;
  email?: string;
  // subject is the `sub` claim; the PDP identifies the principal as `user::<sub>` (ADR 0001).
  subject?: string;
  issuer?: string;
  // expiresAt is the session expiry as unix seconds, as reported by oidc-client-ts.
  expiresAt?: number;
  roles: string[];
}

// useProfile gathers everything the header and the profile page show about the signed-in
// user. Name and contact details come from the ID token; roles only exist in the access
// token (see ADR 0002 and the `roles` protocol mapper in the shipped Keycloak realm).
export const useProfile = (): UserProfile => {
  const auth = useAuth();
  const profile = auth.user?.profile;
  const name = displayName(profile);

  return {
    displayName: name,
    initials: initials(name),
    email: profile?.email,
    subject: profile?.sub,
    issuer: profile?.iss,
    expiresAt: auth.user?.expires_at,
    roles: rolesFromToken(auth.user?.access_token),
  };
};
