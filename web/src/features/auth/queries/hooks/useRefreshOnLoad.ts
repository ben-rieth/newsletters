import { useEffect } from 'react';
import client from '#/api/client';
import { clearSession, dispatchAuthChange } from '#/features/auth/lib/session';

const getTokenExpiry = (token: string): number => {
  const payload = JSON.parse(atob(token.split('.')[1]));
  return payload.exp * 1000;
};

export const useRefreshOnLoad = () => {
  useEffect(() => {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      return;
    }

    const token = localStorage.getItem('token');
    if (token) {
      const expiresAt = getTokenExpiry(token);
      if (Date.now() < expiresAt - 60_000) return;
    }

    const refresh = async () => {
      const { data, error } = await client.POST('/auth/refresh', {
        headers: { Authorization: `Bearer ${refreshToken}` },
      });

      if (error) {
        clearSession();
        return;
      }

      localStorage.setItem('token', data.tokens.token);
      localStorage.setItem('refreshToken', data.tokens.refreshToken);
      dispatchAuthChange();
    };

    refresh();
  }, []);
};
