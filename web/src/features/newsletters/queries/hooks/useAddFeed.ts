import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedsKeys } from '../feeds';
import type { components } from '#/api/schema';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useAddFeed = (newsletterId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      body: components['schemas']['SubmittableFeedFields'],
    ) => {
      const { error } = await client.POST('/newsletter/{newsletterId}/feed', {
        params: { path: { newsletterId } },
        body,
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: feedsKeys.list(newsletterId),
      });
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useAddFeed;
