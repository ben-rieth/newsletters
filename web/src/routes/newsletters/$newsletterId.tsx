import { createFileRoute, Outlet } from '@tanstack/react-router';
import { RouteNotFoundComponent } from '#/components/RouteBoundaries';
import { newsletterOptions } from '#/features/newsletters/queries/newsletters';
import { feedsOptions } from '#/features/newsletters/queries/feeds';
import { feedImportsOptions } from '#/features/newsletters/queries/feedImports';
import { issuesOptions } from '#/features/issues/queries/issues';

export const Route = createFileRoute('/newsletters/$newsletterId')({
  component: () => <Outlet />,
  loader: async ({ context, params }) => {
    await Promise.all([
      context.queryClient.query({
        ...newsletterOptions(params.newsletterId),
        staleTime: 'static',
      }),
      context.queryClient.query({
        ...feedsOptions(params.newsletterId),
        staleTime: 'static',
      }),
      context.queryClient.query({ ...issuesOptions, staleTime: 'static' }),
      context.queryClient.query({
        ...feedImportsOptions(params.newsletterId),
        staleTime: 'static',
      }),
    ]);
  },
  notFoundComponent: () => (
    <RouteNotFoundComponent
      title="Newsletter not found"
      description="It may have been deleted, or the link is stale."
    />
  ),
});
