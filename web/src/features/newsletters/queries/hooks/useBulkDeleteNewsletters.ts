import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { removeDeletedNewsletters } from '../newsletters';
import client from '#/api/client';
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
      removeDeletedNewsletters(queryClient, ids);
      onSuccess?.(ids);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useBulkDeleteNewsletters;
