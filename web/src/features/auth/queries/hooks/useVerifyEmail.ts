import { useMutation } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import client from '#/api/client';
import {
  clearPendingVerificationPassword,
  getPendingVerificationPassword,
} from '#/features/auth/lib/pendingVerificationPassword';
import { dispatchAuthChange } from '#/features/auth/lib/session';

const useVerifyEmail = () => {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (code: string) => {
      const email = localStorage.getItem('pendingVerificationEmail') ?? '';
      const password = getPendingVerificationPassword() ?? '';
      const { data, error } = await client.POST('/auth/verify', {
        body: { code, email, password },
      });
      if (error) {
        throw error;
      }
      return data;
    },
    onSuccess: () => {
      localStorage.removeItem('pendingVerificationEmail');
      clearPendingVerificationPassword();
      dispatchAuthChange();
      navigate({ to: '/issues' });
    },
  });
};

export default useVerifyEmail;
