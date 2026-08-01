import { useSyncExternalStore } from 'react';
import {
  subscribeToAuthChanges,
  getIsSignedIn,
} from '#/features/auth/lib/session';

const useIsSignedIn = () => {
  return useSyncExternalStore(subscribeToAuthChanges, getIsSignedIn);
};

export default useIsSignedIn;
