import { useEffect } from 'react';
import client from '#/api/client';
import {
  getIsSignedIn,
  clearSession,
  dispatchAuthChange,
} from '#/features/auth/lib/session';

export const useRefreshOnLoad = () => {
  useEffect(() => {
    if (!getIsSignedIn()) return;

    const refresh = async () => {
      const { error } = await client.POST('/auth/refresh');

      if (error) {
        clearSession();
        return;
      }

      dispatchAuthChange();
    };

    refresh();
  }, []);
};
