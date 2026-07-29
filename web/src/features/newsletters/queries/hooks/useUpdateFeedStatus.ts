import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedsKeys, feedDetailKeys } from '../feeds';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useUpdateFeedStatus = (
  newsletterId: string,
  onSuccess?: (status: 'active' | 'inactive') => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      feedId,
      status,
    }: {
      feedId: string;
      status: 'active' | 'inactive';
    }) => {
      const { error } = await client.PATCH(
        '/newsletter/{newsletterId}/feed/{feedId}/status',
        {
          params: { path: { newsletterId, feedId } },
          body: { status },
        },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: async (_data, { feedId, status }) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: feedsKeys.list(newsletterId),
        }),
        queryClient.invalidateQueries({
          queryKey: feedDetailKeys.detail(newsletterId, feedId),
        }),
      ]);
      onSuccess?.(status);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useUpdateFeedStatus;
