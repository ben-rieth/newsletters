import { useMutation } from '@tanstack/react-query';
import client from '#/api/client';

const useResendEmailUpdateVerification = () => {
  return useMutation({
    mutationFn: async () => {
      const { error } = await client.POST('/user/verify/resend', {});
      if (error) {
        throw error;
      }
    },
  });
};

export default useResendEmailUpdateVerification;
