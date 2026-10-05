import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { userKeys } from '#/features/auth/queries/user';
import type { VisibleUser } from '#/features/auth/queries/user';
import { issueKeys } from '../issues';

const useUpdateIssueRetention = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (issueRetentionDays: number) => {
      const { error } = await client.PATCH('/user/issue-retention', {
        body: { issueRetentionDays },
      });
      if (error) {
        throw error;
      }
    },
    onMutate: async (issueRetentionDays) => {
      await queryClient.cancelQueries({ queryKey: userKeys.me });

      const previousUser = queryClient.getQueryData<VisibleUser>(userKeys.me);

      queryClient.setQueryData<VisibleUser>(
        userKeys.me,
        (user) => user && { ...user, issueRetentionDays },
      );

      return { previousUser };
    },
    onSuccess: () => {
      // Shortening the window can take issues with it on the next prune.
      queryClient.invalidateQueries({ queryKey: issueKeys.all, exact: true });
    },
    onError: (error, _days, context) => {
      if (context?.previousUser) {
        queryClient.setQueryData(userKeys.me, context.previousUser);
      }
      toast.error(getErrorMessage(error));
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.me });
    },
  });
};

export default useUpdateIssueRetention;
