import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedDetailKeys } from '../feeds';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useDeleteFeedFilter = (
  newsletterId: string,
  feedId: string,
  onSuccess?: () => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (filterId: string) => {
      const { error } = await client.DELETE(
        '/newsletter/{newsletterId}/feed/{feedId}/filter/{filterId}',
        { params: { path: { newsletterId, feedId, filterId } } },
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

export default useDeleteFeedFilter;
