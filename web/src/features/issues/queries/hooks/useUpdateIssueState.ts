import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { issueKeys } from '../issues';
import type { DetailedIssue, Issue, IssueState } from '../issues';

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
      await Promise.all([
        queryClient.cancelQueries({ queryKey: issueKeys.all }),
        queryClient.cancelQueries({ queryKey: issueKeys.detail(issueId) }),
      ]);

      const previousList = queryClient.getQueryData<Issue[]>(issueKeys.all);
      const previousDetail = queryClient.getQueryData<DetailedIssue>(
        issueKeys.detail(issueId),
      );

      queryClient.setQueryData<Issue[]>(issueKeys.all, (issues) =>
        issues?.map((issue) =>
          issue.issueId === issueId ? { ...issue, state } : issue,
        ),
      );
      queryClient.setQueryData<DetailedIssue>(
        issueKeys.detail(issueId),
        (issue) =>
          issue
            ? {
                ...issue,
                state,
                feeds: (issue.feeds ?? []).map((feed) => ({
                  ...feed,
                  items: (feed.items ?? []).map((item) => ({ ...item, state })),
                })),
              }
            : issue,
      );

      return { previousList, previousDetail };
    },
    onError: (error, _state, context) => {
      if (context?.previousList) {
        queryClient.setQueryData(issueKeys.all, context.previousList);
      }
      if (context?.previousDetail) {
        queryClient.setQueryData(
          issueKeys.detail(issueId),
          context.previousDetail,
        );
      }
      toast.error(getErrorMessage(error));
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: issueKeys.all });
      queryClient.invalidateQueries({ queryKey: issueKeys.detail(issueId) });
    },
  });
};

export default useUpdateIssueState;
