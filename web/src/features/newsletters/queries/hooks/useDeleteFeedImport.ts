import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedImportsKeys } from '../feedImports';
import { newslettersKeys } from '../newsletters';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useDeleteFeedImport = (newsletterId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (importId: string) => {
      const { error } = await client.DELETE(
        '/newsletter/{newsletterId}/feed-imports/{importId}',
        { params: { path: { newsletterId, importId } } },
      );
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: feedImportsKeys.list(newsletterId),
        }),
        queryClient.invalidateQueries({ queryKey: newslettersKeys.all }),
      ]);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useDeleteFeedImport;
