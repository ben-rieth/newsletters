import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useUnsubscribe = (onSuccess?: () => void) => {
  return useMutation({
    mutationFn: async (unsubscribeToken: string) => {
      const { error } = await client.POST('/unsubscribe', {
        params: { query: { unsubscribeToken } },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: () => {
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useUnsubscribe;