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
  document.cookie.split(';').some((c) => c.trim().startsWith('signed_in='));

export const clearSession = () => {
  dispatchAuthChange();
  router.navigate({ to: '/sign-in' });
};
