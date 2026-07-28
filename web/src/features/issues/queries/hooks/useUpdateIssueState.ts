import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import {
  applyItemStates,
  invalidateIssueCaches,
  restoreIssueCaches,
  snapshotIssueCaches,
} from '../issueCache';
import type { IssueState } from '../issues';

const useUpdateIssueState = (issueId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (state: IssueState) => {
      const { error } = await client.PUT('/issues/{issueId}/state', {
        params: { path: { issueId } },
        body: { state },
      });
      if (error) {
        throw error;
      }
    },
    onMutate: async (state) => {
      const snapshot = await snapshotIssueCaches(queryClient, issueId);
      applyItemStates(queryClient, issueId, () => state);
      return snapshot;
    },
    onError: (error, _state, snapshot) => {
      restoreIssueCaches(queryClient, issueId, snapshot);
      toast.error(getErrorMessage(error));
    },
    onSettled: () => invalidateIssueCaches(queryClient, issueId),
  });
};

export default useUpdateIssueState;
