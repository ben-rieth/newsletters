import { router } from '#/router';

export const dispatchAuthChange = () => {
  window.dispatchEvent(new Event('auth-change'));
};

export const subscribeToAuthChanges = (callback: () => void) => {
  window.addEventListener('auth-change', callback);
  return () => {
    window.removeEventListener('auth-change', callback);
  };
};

export const getIsSignedIn = () =>
  document.cookie
    .split(';')
    .some((c) => c.trim().startsWith('__Host-signed_in='));

export const clearSession = () => {
  // Only an explicit revoke clears this server side, so a failed refresh would
  // otherwise leave it set for its full 30 days and keep getIsSignedIn() lying.
  document.cookie =
    '__Host-signed_in=; Path=/; Max-Age=0; Secure; SameSite=Strict';

  dispatchAuthChange();
  router.navigate({ to: '/sign-in' });
};
