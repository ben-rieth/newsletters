import { useMutation } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import client from '#/api/client';
import { dispatchAuthChange } from '#/features/auth/lib/session';

const useVerifyEmail = () => {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (code: string) => {
      const email = localStorage.getItem('pendingVerificationEmail') ?? '';
      const { data, error } = await client.POST('/auth/verify', {
        body: { code, email },
      });
      if (error) {
        throw error;
      }
      return data;
    },
    onSuccess: () => {
      localStorage.removeItem('pendingVerificationEmail');
      dispatchAuthChange();
      navigate({ to: '/newsletters' });
    },
  });
};

export default useVerifyEmail;
