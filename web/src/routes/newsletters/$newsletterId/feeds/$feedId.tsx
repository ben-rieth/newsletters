import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { H2 } from '#/components/ui/typography'
import { FeedDetail } from '#/features/newsletters/components/FeedDetail'
import { feedDetailOptions } from '#/features/newsletters/queries/feeds'
import { newsletterOptions } from '#/features/newsletters/queries/newsletters'

const FeedDetailPage = () => {
  const { newsletterId, feedId } = Route.useParams()
  const { data: newsletter } = useSuspenseQuery(newsletterOptions(newsletterId))
  const { data: feed } = useSuspenseQuery(feedDetailOptions(newsletterId, feedId))

  return (
    <div className="p-6 space-y-6">
      <div>
        <Link
          to="/newsletters/$newsletterId"
          params={{ newsletterId }}
          className="text-sm text-muted-foreground hover:underline"
        >
          ← {newsletter.name}
        </Link>
        <H2 className="mt-2">{feed.alias || feed.title}</H2>
      </div>

      <FeedDetail newsletterId={newsletterId} feedId={feedId} />
    </div>
  )
}

export const Route = createFileRoute('/newsletters/$newsletterId/feeds/$feedId')({
  component: FeedDetailPage,
  loader: async ({ context, params }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(newsletterOptions(params.newsletterId)),
      context.queryClient.ensureQueryData(
        feedDetailOptions(params.newsletterId, params.feedId),
      ),
    ])
  },
})
