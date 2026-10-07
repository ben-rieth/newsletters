import client from '#/api/client';
import type { components } from '#/api/schema';
import { queryOptions } from '@tanstack/react-query';
import { feedsKeys } from './feeds';

export type FeedImport = components['schemas']['FeedImport'];

export const feedImportsKeys = {
  list: (newsletterId: string) => ['feedImports', newsletterId] as const,
};

const PENDING_POLL_MS = 3000;

const pendingIds = (imports: FeedImport[] | undefined) =>
  new Set(
    (imports ?? []).filter((i) => i.state === 'pending').map((i) => i.id),
  );

export const feedImportsOptions = (newsletterId: string) =>
  queryOptions({
    queryKey: feedImportsKeys.list(newsletterId),
    queryFn: async ({ client: queryClient }) => {
      const { data, error } = await client.GET(
        '/newsletter/{newsletterId}/feed-imports',
        { params: { path: { newsletterId } } },
      );
      if (error) {
        throw error;
      }

      const imports = data ?? [];
      const stillPending = pendingIds(imports);
      const previouslyPending = pendingIds(
        queryClient.getQueryData(feedImportsKeys.list(newsletterId)),
      );
      // A pending import that vanished has become a real feed.
      if ([...previouslyPending].some((id) => !stillPending.has(id))) {
        await queryClient.invalidateQueries({
          queryKey: feedsKeys.list(newsletterId),
        });
      }

      return imports;
    },
    refetchInterval: (query) =>
      pendingIds(query.state.data).size > 0 ? PENDING_POLL_MS : false,
  });
