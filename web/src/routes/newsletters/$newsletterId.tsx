import { createFileRoute, Outlet } from '@tanstack/react-router'
import { newsletterOptions } from '#/features/newsletters/queries/newsletters'
import { feedsOptions } from '#/features/newsletters/queries/feeds'

export const Route = createFileRoute('/newsletters/$newsletterId')({
  component: () => <Outlet />,
  loader: async ({ context, params }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(newsletterOptions(params.newsletterId)),
      context.queryClient.ensureQueryData(feedsOptions(params.newsletterId)),
    ])
  },
  notFoundComponent: () => (
    <div className="p-6">
      <p className="text-muted-foreground">Newsletter not found.</p>
    </div>
  ),
})
