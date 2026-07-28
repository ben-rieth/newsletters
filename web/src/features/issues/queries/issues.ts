import client from '#/api/client';
import type { components } from '#/api/schema';
import { queryOptions } from '@tanstack/react-query';

export type Issue = components['schemas']['Issue'];
export type DetailedIssue = components['schemas']['DetailedIssue'];
export type IssueFeed = components['schemas']['IssueFeed'];
export type IssueItem = components['schemas']['IssueItem'];

export type IssueState = Issue['state'];

export const issueKeys = {
  all: ['issues' as const],
  detail: (id: string) => ['issues', id] as const,
  itemState: (id: string) => ['issues', id, 'itemState'] as const,
};

export const issuesOptions = queryOptions({
  queryKey: issueKeys.all,
  queryFn: async () => {
    const { data, error } = await client.GET('/issues');
    if (error) throw error;
    return data;
  },
});

export const issueDetailOptions = (issueId: string) =>
  queryOptions({
    queryKey: issueKeys.detail(issueId),
    queryFn: async () => {
      const { data, error } = await client.GET('/issues/{issueId}', {
        params: { path: { issueId } },
      });
      if (error) throw error;
      return data;
    },
  });
