import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedDetailKeys } from '../feeds';
import type { components } from '#/api/schema';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useAddFeedFilter = (
  newsletterId: string,
  feedId: string,
  onSuccess?: () => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      body: components['schemas']['SubmittableFeedFilterFields'],
    ) => {
      const { error } = await client.POST(
        '/newsletter/{newsletterId}/feed/{feedId}/filter',
        { params: { path: { newsletterId, feedId } }, body },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: feedDetailKeys.detail(newsletterId, feedId),
      });
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useAddFeedFilter;
