import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { issueKeys } from '../issues';

const useDeleteIssue = (issueId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const { error } = await client.DELETE('/issues/{issueId}', {
        params: { path: { issueId } },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      // The detail query is removed rather than invalidated — refetching an issue
      // that no longer exists would only surface an error on the way out.
      queryClient.removeQueries({ queryKey: issueKeys.detail(issueId) });
      await queryClient.invalidateQueries({
        queryKey: issueKeys.all,
        exact: true,
      });
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useDeleteIssue;
