import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedDetailKeys } from '../feeds';
import type { components } from '#/api/schema';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

type UpdateFilterArgs = {
  filterId: string;
  body: components['schemas']['SubmittableFeedFilterFields'];
};

const useUpdateFeedFilter = (
  newsletterId: string,
  feedId: string,
  onSuccess?: () => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ filterId, body }: UpdateFilterArgs) => {
      const { error } = await client.POST(
        '/newsletter/{newsletterId}/feed/{feedId}/filter/{filterId}',
        { params: { path: { newsletterId, feedId, filterId } }, body },
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

export default useUpdateFeedFilter;
