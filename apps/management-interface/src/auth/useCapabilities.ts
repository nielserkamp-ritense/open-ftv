import { useAuth } from 'react-oidc-context';
import { rolesFromToken, canWrite, canPublish } from './capabilities';

// useCapabilities reads the current user's roles from the access token and exposes the
// UI capability flags. Roles live in the ACCESS token (not the ID token / profile).
export const useCapabilities = (): { roles: string[]; canWrite: boolean; canPublish: boolean } => {
  const auth = useAuth();
  const roles = rolesFromToken(auth.user?.access_token);
  return { roles, canWrite: canWrite(roles), canPublish: canPublish(roles) };
};
