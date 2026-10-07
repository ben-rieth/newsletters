import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedImportsKeys } from '../feedImports';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useRetryFeedImport = (newsletterId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (importId: string) => {
      const { error } = await client.POST(
        '/newsletter/{newsletterId}/feed-imports/{importId}/retry',
        { params: { path: { newsletterId, importId } } },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: feedImportsKeys.list(newsletterId),
      });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useRetryFeedImport;
