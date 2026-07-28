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
import { issueKeys } from '../issues';
import type { IssueState } from '../issues';

type Variables = {
  itemId: string;
  state: IssueState;
};

const useUpdateIssueItemState = (issueId: string) => {
  const queryClient = useQueryClient();
  const mutationKey = issueKeys.itemState(issueId);

  return useMutation({
    mutationKey,
    mutationFn: async ({ itemId, state }: Variables) => {
      const { error } = await client.PUT(
        '/issues/{issueId}/item/{itemId}/state',
        {
          params: { path: { issueId, itemId } },
          body: { state },
        },
      );
      if (error) {
        throw error;
      }
    },
    onMutate: async ({ itemId, state }) => {
      const snapshot = await snapshotIssueCaches(queryClient, issueId);
      applyItemStates(queryClient, issueId, (item) =>
        item.itemId === itemId ? state : item.state,
      );
      return snapshot;
    },
    onError: (error, _variables, snapshot) => {
      restoreIssueCaches(queryClient, issueId, snapshot);
      toast.error(getErrorMessage(error));
    },
    onSettled: () => {
      if (queryClient.isMutating({ mutationKey }) > 1) return;

      invalidateIssueCaches(queryClient, issueId);
    },
  });
};

export default useUpdateIssueItemState;
