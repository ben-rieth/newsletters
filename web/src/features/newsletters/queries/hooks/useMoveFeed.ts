import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedsKeys } from '../feeds';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useMoveFeed = (
  newsletterId: string,
  feedId: string,
  onSuccess?: (targetNewsletterId: string) => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (targetNewsletterId: string) => {
      const { error } = await client.POST(
        '/newsletter/{newsletterId}/feed/{feedId}/move',
        {
          params: { path: { newsletterId, feedId } },
          body: { newsletterId: targetNewsletterId },
        },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: async (_data, targetNewsletterId) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: feedsKeys.list(newsletterId),
        }),
        queryClient.invalidateQueries({
          queryKey: feedsKeys.list(targetNewsletterId),
        }),
      ]);
      onSuccess?.(targetNewsletterId);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useMoveFeed;
