import { useEffect } from 'react';
import { refreshSession } from '#/api/client';
import {
  getIsSignedIn,
  clearSession,
  dispatchAuthChange,
} from '#/features/auth/lib/session';

export const useRefreshOnLoad = () => {
  useEffect(() => {
    if (!getIsSignedIn()) return;

    const refresh = async () => {
      const refreshed = await refreshSession();

      if (!refreshed) {
        clearSession();
        return;
      }

      dispatchAuthChange();
    };

    refresh();
  }, []);
};
