import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { issueKeys } from '../issues';
import type { DetailedIssue, Issue, IssueState } from '../issues';

type Variables = {
  itemId: string;
  state: IssueState;
};

const rollUpState = (issue: DetailedIssue): IssueState =>
  (issue.feeds ?? []).some((feed) =>
    (feed.items ?? []).some((item) => item.state === 'unread'),
  )
    ? 'unread'
    : 'read';

const useUpdateIssueItemState = (issueId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
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
      await Promise.all([
        queryClient.cancelQueries({ queryKey: issueKeys.all }),
        queryClient.cancelQueries({ queryKey: issueKeys.detail(issueId) }),
      ]);

      const previousList = queryClient.getQueryData<Issue[]>(issueKeys.all);
      const previousDetail = queryClient.getQueryData<DetailedIssue>(
        issueKeys.detail(issueId),
      );

      const nextDetail = queryClient.setQueryData<DetailedIssue>(
        issueKeys.detail(issueId),
        (issue) => {
          if (!issue) return issue;
          const withItem = {
            ...issue,
            feeds: (issue.feeds ?? []).map((feed) => ({
              ...feed,
              items: (feed.items ?? []).map((item) =>
                item.itemId === itemId ? { ...item, state } : item,
              ),
            })),
          };
          return { ...withItem, state: rollUpState(withItem) };
        },
      );

      if (nextDetail) {
        queryClient.setQueryData<Issue[]>(issueKeys.all, (issues) =>
          issues?.map((issue) =>
            issue.issueId === issueId
              ? { ...issue, state: nextDetail.state }
              : issue,
          ),
        );
      }

      return { previousList, previousDetail };
    },
    onError: (error, _variables, context) => {
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

export default useUpdateIssueItemState;
