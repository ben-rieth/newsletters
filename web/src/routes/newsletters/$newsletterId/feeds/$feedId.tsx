import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { ChevronLeft } from 'lucide-react';
import { z } from 'zod';
import {
  FEED_TABS,
  FeedDetail,
} from '#/features/newsletters/components/FeedDetail';
import { feedDetailOptions } from '#/features/newsletters/queries/feeds';
import {
  newsletterOptions,
  newslettersOptions,
} from '#/features/newsletters/queries/newsletters';
import { useMobileHeader } from '#/components/MobileHeader';

const FeedDetailPage = () => {
  const { newsletterId, feedId } = Route.useParams();
  const { tab = 'details' } = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: newsletter } = useSuspenseQuery(
    newsletterOptions(newsletterId),
  );
  const { data: feed } = useSuspenseQuery(
    feedDetailOptions(newsletterId, feedId),
  );

  useMobileHeader({
    title: feed.alias || feed.title,
    back: {
      to: '/newsletters/$newsletterId',
      params: { newsletterId },
    },
  });

  return (
    <div className="mx-auto max-w-5xl px-4 py-6 md:px-6 md:py-8 lg:px-10">
      <Link
        to="/newsletters/$newsletterId"
        params={{ newsletterId }}
        className="mb-6 hidden items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground md:inline-flex"
      >
        <ChevronLeft className="size-4" />
        {newsletter.name}
      </Link>

      <FeedDetail
        newsletterId={newsletterId}
        feedId={feedId}
        tab={tab}
        onTabChange={(next) =>
          navigate({
            search: { tab: next === 'details' ? undefined : next },
            replace: true,
            resetScroll: false,
          })
        }
      />
    </div>
  );
};

export const Route = createFileRoute(
  '/newsletters/$newsletterId/feeds/$feedId',
)({
  component: FeedDetailPage,
  validateSearch: z.object({
    tab: z.enum(FEED_TABS).optional().catch(undefined),
  }),
  loader: async ({ context, params }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(
        newsletterOptions(params.newsletterId),
      ),
      context.queryClient.ensureQueryData(newslettersOptions),
      context.queryClient.ensureQueryData(
        feedDetailOptions(params.newsletterId, params.feedId),
      ),
    ]);
  },
});
