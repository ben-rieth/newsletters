import { useEffect } from 'react';
import { clearSession, dispatchAuthChange } from '#/features/auth/lib/session';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export const useRefreshOnLoad = () => {
  useEffect(() => {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) { return; }

    const refresh = async () => {
      const response = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${refreshToken}` },
      });

      if (!response.ok) {
        clearSession();
        return;
      }

      const data = await response.json() as { token: string; refreshToken: string };
      localStorage.setItem('token', data.token);
      localStorage.setItem('refreshToken', data.refreshToken);
      dispatchAuthChange();
    };

    refresh();
  }, []);
};
