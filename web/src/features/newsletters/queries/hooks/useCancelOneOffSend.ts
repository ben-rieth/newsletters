import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { newslettersKeys } from '../newsletters';
import { getErrorMessage } from '#/lib/errors';

const useCancelOneOffSend = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (newsletterId: string) => {
      const { error } = await client.DELETE(
        '/newsletter/{newsletterId}/one-off',
        {
          params: { path: { newsletterId } },
        },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: (_data, newsletterId) => {
      queryClient.invalidateQueries({
        queryKey: newslettersKeys.detail(newsletterId),
      });
      queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useCancelOneOffSend;
