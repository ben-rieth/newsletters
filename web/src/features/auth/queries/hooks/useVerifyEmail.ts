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
    onSuccess: (data) => {
      localStorage.removeItem('pendingVerificationEmail');
      localStorage.setItem('token', data.tokens.token);
      localStorage.setItem('refreshToken', data.tokens.refreshToken);
      dispatchAuthChange();
      navigate({ to: '/newsletters' });
    },
  });
};

export default useVerifyEmail;
