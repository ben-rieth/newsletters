import type { QueryClient } from '@tanstack/react-query';
import { issueKeys } from './issues';
import type { DetailedIssue, Issue, IssueItem, IssueState } from './issues';

export type IssueCacheSnapshot = {
  previousList: Issue[] | undefined;
  previousDetail: DetailedIssue | undefined;
};

const rollUpState = (issue: DetailedIssue): IssueState =>
  (issue.feeds ?? []).some((feed) =>
    (feed.items ?? []).some((item) => item.state === 'unread'),
  )
    ? 'unread'
    : 'read';

export const snapshotIssueCaches = async (
  queryClient: QueryClient,
  issueId: string,
): Promise<IssueCacheSnapshot> => {
  await Promise.all([
    queryClient.cancelQueries({ queryKey: issueKeys.all, exact: true }),
    queryClient.cancelQueries({ queryKey: issueKeys.detail(issueId) }),
  ]);

  return {
    previousList: queryClient.getQueryData<Issue[]>(issueKeys.all),
    previousDetail: queryClient.getQueryData<DetailedIssue>(
      issueKeys.detail(issueId),
    ),
  };
};

export const applyItemStates = (
  queryClient: QueryClient,
  issueId: string,
  nextState: (item: IssueItem) => IssueState,
) => {
  const nextDetail = queryClient.setQueryData<DetailedIssue>(
    issueKeys.detail(issueId),
    (issue) => {
      if (!issue) return issue;
      const withItems = {
        ...issue,
        feeds: (issue.feeds ?? []).map((feed) => ({
          ...feed,
          items: (feed.items ?? []).map((item) => ({
            ...item,
            state: nextState(item),
          })),
        })),
      };
      return { ...withItems, state: rollUpState(withItems) };
    },
  );

  if (!nextDetail) return;

  queryClient.setQueryData<Issue[]>(issueKeys.all, (issues) =>
    issues?.map((issue) =>
      issue.issueId === issueId ? { ...issue, state: nextDetail.state } : issue,
    ),
  );
};

export const restoreIssueCaches = (
  queryClient: QueryClient,
  issueId: string,
  snapshot: IssueCacheSnapshot | undefined,
) => {
  if (snapshot?.previousList) {
    queryClient.setQueryData(issueKeys.all, snapshot.previousList);
  }
  if (snapshot?.previousDetail) {
    queryClient.setQueryData(
      issueKeys.detail(issueId),
      snapshot.previousDetail,
    );
  }
};

export const invalidateIssueCaches = (
  queryClient: QueryClient,
  issueId: string,
) => {
  queryClient.invalidateQueries({ queryKey: issueKeys.all, exact: true });
  queryClient.invalidateQueries({ queryKey: issueKeys.detail(issueId) });
};
