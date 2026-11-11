import { clearPendingVerificationPassword } from '#/features/auth/lib/pendingVerificationPassword';
import { queryClient, router } from '#/router';

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

  clearPendingVerificationPassword();
  dispatchAuthChange();
  // Clearing while the signed-in page is still mounted would make its
  // suspense queries refetch without a session and 401 straight back here.
  void router.navigate({ to: '/sign-in' }).then(() => queryClient.clear());
};
