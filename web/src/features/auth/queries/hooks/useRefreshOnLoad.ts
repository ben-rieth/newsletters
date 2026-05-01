import { useEffect } from 'react';
import client from '#/api/client';
import { clearSession, dispatchAuthChange } from '#/features/auth/lib/session';

export const useRefreshOnLoad = () => {
  useEffect(() => {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      return;
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
