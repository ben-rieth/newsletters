import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { newslettersKeys } from '../newsletters';
import { feedsKeys } from '../feeds';
import { feedImportsKeys } from '../feedImports';
import client from '#/api/client';
import { issueKeys } from '#/features/issues/queries/issues';
import { getErrorMessage } from '#/lib/errors';

const useBulkDeleteNewsletters = (onSuccess?: (ids: string[]) => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      const { error } = await client.POST('/newsletters/bulk-delete', {
        body: { ids },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: (_data, ids) => {
      for (const id of ids) {
        queryClient.removeQueries({ queryKey: newslettersKeys.detail(id) });
        queryClient.removeQueries({ queryKey: feedsKeys.list(id) });
        queryClient.removeQueries({ queryKey: feedImportsKeys.list(id) });
      }
      queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
      queryClient.invalidateQueries({ queryKey: issueKeys.all });
      onSuccess?.(ids);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useBulkDeleteNewsletters;
