import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { clearSession } from '#/features/auth/lib/session';

const useDeleteAccount = () => {
  return useMutation({
    mutationFn: async ({ password }: { password: string }) => {
      const { error } = await client.DELETE('/user', {
        body: { password },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: () => {
      clearSession();
      toast.success('Your user account and all connected data was deleted.');
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useDeleteAccount;
