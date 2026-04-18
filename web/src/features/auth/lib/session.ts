import { router } from '#/router';

export const dispatchAuthChange = () => {
  window.dispatchEvent(new Event('auth-change'));
};

export const subscribeToAuthChanges = (callback: () => void) => {
  window.addEventListener('auth-change', callback);
  window.addEventListener('storage', callback);
  return () => {
    window.removeEventListener('auth-change', callback);
    window.removeEventListener('storage', callback);
  };
};

export const getIsSignedIn = () => !!localStorage.getItem('token');

export const clearSession = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('refreshToken');
  dispatchAuthChange();
  router.navigate({ to: '/sign-in' });
};