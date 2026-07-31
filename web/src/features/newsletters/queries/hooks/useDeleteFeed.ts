import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedsKeys } from '../feeds';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useDeleteFeed = (newsletterId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (feedId: string) => {
      const { error } = await client.DELETE(
        '/newsletter/{newsletterId}/feed/{feedId}',
        { params: { path: { newsletterId, feedId } } },
      );
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

export default useDeleteFeed;
