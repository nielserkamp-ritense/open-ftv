import type { AuthProviderProps } from 'react-oidc-context';
import { WebStorageStateStore } from 'oidc-client-ts';
import { OIDC_AUTHORITY, OIDC_CLIENT_ID } from '@/config/env';

export const oidcConfig: AuthProviderProps = {
  authority: OIDC_AUTHORITY,
  client_id: OIDC_CLIENT_ID,
  redirect_uri: window.location.origin + '/',
  post_logout_redirect_uri: window.location.origin + '/',
  response_type: 'code',
  scope: 'openid profile email',
  userStore: new WebStorageStateStore({ store: window.localStorage }),
  // The login keeps the page it started on as its state, so a link into the management interface still opens
  // its page after logging in.
  onSigninCallback: (user) => {
    const page = typeof user?.state === 'string' ? user.state : window.location.pathname;
    window.history.replaceState({}, document.title, page);
  },
};
