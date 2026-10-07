import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { newslettersKeys } from '../newsletters';
import { feedsKeys } from '../feeds';
import { feedImportsKeys } from '../feedImports';
import client from '#/api/client';
import { issueKeys } from '#/features/issues/queries/issues';
import { getErrorMessage } from '#/lib/errors';

const useDeleteNewsletter = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (newsletterId: string) => {
      const { error } = await client.DELETE('/newsletters/{newsletterId}', {
        params: { path: { newsletterId } },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: (_, newsletterId) => {
      // Invalidating would refetch the still-mounted detail queries, which now
      // 404 and sit through the retry backoff before navigation can happen.
      queryClient.removeQueries({
        queryKey: newslettersKeys.detail(newsletterId),
      });
      queryClient.removeQueries({ queryKey: feedsKeys.list(newsletterId) });
      queryClient.removeQueries({
        queryKey: feedImportsKeys.list(newsletterId),
      });
      queryClient.invalidateQueries({
        queryKey: newslettersKeys.all,
        exact: true,
      });
      queryClient.invalidateQueries({ queryKey: issueKeys.all });
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useDeleteNewsletter;
