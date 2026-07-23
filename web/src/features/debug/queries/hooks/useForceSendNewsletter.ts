import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { newslettersKeys } from '#/features/newsletters/queries/newsletters';
import { getErrorMessage } from '#/lib/errors';

/**
 * Debug-only: immediately queues a newsletter to send. Backed by an endpoint
 * that is only registered when the server runs in a dev environment.
 */
const useForceSendNewsletter = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (newsletterId: string) => {
      const { error } = await client.POST(
        '/debug/newsletters/{newsletterId}/send',
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

export default useForceSendNewsletter;
