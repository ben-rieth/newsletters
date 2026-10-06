import {
  useMutation,
  useMutationState,
  useQueryClient,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedsKeys, feedDetailKeys } from '../feeds';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

type Variables = {
  feedId: string;
  status: 'active' | 'inactive';
};

export const usePendingFeedStatusIds = (newsletterId: string) =>
  new Set(
    useMutationState({
      filters: {
        mutationKey: feedsKeys.status(newsletterId),
        status: 'pending',
      },
      select: (mutation) => (mutation.state.variables as Variables).feedId,
    }),
  );

const useUpdateFeedStatus = (
  newsletterId: string,
  onSuccess?: (status: 'active' | 'inactive') => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationKey: feedsKeys.status(newsletterId),
    mutationFn: async ({ feedId, status }: Variables) => {
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
