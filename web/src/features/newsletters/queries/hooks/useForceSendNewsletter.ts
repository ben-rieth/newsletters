import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import client from '#/api/client'
import { getErrorMessage } from '#/lib/errors'

const useForceSendNewsletter = (onSuccess?: () => void) => {
  return useMutation({
    mutationFn: async (newsletterId: string) => {
      const { error } = await client.PATCH('/newsletter/{newsletterId}/send', {
        params: { path: { newsletterId } },
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

export default useForceSendNewsletter;
