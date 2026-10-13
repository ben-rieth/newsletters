import { useMutation } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import client from '#/api/client';
import type { components } from '#/api/schema';
import {
  clearPendingVerificationPassword,
  setPendingVerificationPassword,
} from '#/features/auth/lib/pendingVerificationPassword';
import { dispatchAuthChange } from '#/features/auth/lib/session';

type AuthPath = '/auth/sign-in' | '/auth/sign-up';

const useAuth = (path: AuthPath) => {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (body: components['schemas']['AuthInputBody']) => {
      const { data, error } = await client.POST(path, { body });
      if (error) {
        throw error;
      }
      return data;
    },
    onSuccess: (data, variables) => {
      if (!data.verified) {
        localStorage.setItem('pendingVerificationEmail', variables.email);
        setPendingVerificationPassword(variables.password);
        navigate({ to: '/verify-email' });
        return;
      }
      clearPendingVerificationPassword();
      dispatchAuthChange();
      navigate({ to: '/issues' });
    },
  });
};

export default useAuth;
