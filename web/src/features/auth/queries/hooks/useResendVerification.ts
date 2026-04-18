import { useMutation } from '@tanstack/react-query';
import client from '#/api/client';

const useResendVerification = () => {
  return useMutation({
    mutationFn: async () => {
      const email = localStorage.getItem('pendingVerificationEmail') ?? '';
      const { error } = await client.POST('/auth/verify/resend', {
        body: { email },
      });
      if (error) {
        throw error;
      }
    },
  });
};

export default useResendVerification;
