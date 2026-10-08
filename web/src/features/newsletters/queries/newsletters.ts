import client from '#/api/client';
import type { components } from '#/api/schema';
import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import { issueKeys } from '#/features/issues/queries/issues';
import { feedsKeys } from './feeds';
import { feedImportsKeys } from './feedImports';

export type Newsletter = components['schemas']['Newsletter'];

export type Frequency = 'daily' | 'weekly' | 'monthly';

export const newslettersKeys = {
  all: ['newsletters' as const],
  detail: (id: string) => ['newsletters', id] as const,
};

// Invalidating would refetch the still-mounted detail queries, which now
// 404 and sit through the retry backoff before navigation can happen.
export const removeDeletedNewsletters = (
  queryClient: QueryClient,
  ids: string[],
) => {
  for (const id of ids) {
    queryClient.removeQueries({ queryKey: newslettersKeys.detail(id) });
    queryClient.removeQueries({ queryKey: feedsKeys.list(id) });
    queryClient.removeQueries({ queryKey: feedImportsKeys.list(id) });
  }
  queryClient.invalidateQueries({ queryKey: newslettersKeys.all, exact: true });
  queryClient.invalidateQueries({ queryKey: issueKeys.all });
};

const PENDING_IMPORTS_POLL_MS = 5000;

const hasPendingImports = (newsletters: Newsletter[] | undefined) =>
  (newsletters ?? []).some((n) => n.pendingFeedImports > 0);

export const newslettersOptions = queryOptions({
  queryKey: newslettersKeys.all,
  queryFn: async () => {
    const { data, error } = await client.GET('/newsletters');
    if (error) {
      throw error;
    }
    return data;
  },
  refetchInterval: (query) =>
    hasPendingImports(query.state.data) ? PENDING_IMPORTS_POLL_MS : false,
});

export const newsletterOptions = (newsletterId: string) => {
  return queryOptions({
    queryKey: newslettersKeys.detail(newsletterId),
    queryFn: async () => {
      const { data, error } = await client.GET('/newsletters/{newsletterId}', {
        params: { path: { newsletterId } },
      });
      if (error) {
        throw error;
      }
      return data;
    },
    refetchInterval: (query) =>
      (query.state.data?.pendingFeedImports ?? 0) > 0
        ? PENDING_IMPORTS_POLL_MS
        : false,
  });
};
