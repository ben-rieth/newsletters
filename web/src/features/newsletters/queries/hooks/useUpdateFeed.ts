import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedsKeys, feedDetailKeys } from '../feeds';
import type { components } from '#/api/schema';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useUpdateFeed = (
  newsletterId: string,
  feedId: string,
  onSuccess?: () => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (body: components['schemas']['UpdatableFeedFields']) => {
      const { error } = await client.PUT(
        '/newsletter/{newsletterId}/feed/{feedId}',
        { params: { path: { newsletterId, feedId } }, body },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: feedsKeys.list(newsletterId),
        }),
        queryClient.invalidateQueries({
          queryKey: feedDetailKeys.detail(newsletterId, feedId),
        }),
      ]);
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useUpdateFeed;
