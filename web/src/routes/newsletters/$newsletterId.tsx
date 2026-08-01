import { createFileRoute, Outlet } from '@tanstack/react-router';
import { RouteNotFoundComponent } from '#/components/RouteBoundaries';
import { newsletterOptions } from '#/features/newsletters/queries/newsletters';
import { feedsOptions } from '#/features/newsletters/queries/feeds';
import { issuesOptions } from '#/features/issues/queries/issues';

export const Route = createFileRoute('/newsletters/$newsletterId')({
  component: () => <Outlet />,
  loader: async ({ context, params }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(
        newsletterOptions(params.newsletterId),
      ),
      context.queryClient.ensureQueryData(feedsOptions(params.newsletterId)),
      context.queryClient.ensureQueryData(issuesOptions),
    ]);
  },
  notFoundComponent: () => (
    <RouteNotFoundComponent
      title="Newsletter not found"
      description="It may have been deleted, or the link is stale."
    />
  ),
});
