import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { clearSession } from '#/features/auth/lib/session';

const useLogout = () => {
  return useMutation({
    mutationFn: async () => {
      const { error } = await client.POST('/auth/revoke');
      if (error) {
        throw error;
      }
    },
    onSuccess: () => {
      clearSession();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useLogout;
