import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { newslettersKeys } from '../newsletters';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

type Variables = { ids: string[]; status: 'active' | 'inactive' };

const useBulkUpdateNewsletterStatus = (
  onSuccess?: (variables: Variables) => void,
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ ids, status }: Variables) => {
      const { error } = await client.PATCH('/newsletters/status', {
        body: { ids, status },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
      onSuccess?.(variables);
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useBulkUpdateNewsletterStatus;
