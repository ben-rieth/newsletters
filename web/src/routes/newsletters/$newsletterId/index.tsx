import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { NewsletterDetail } from '#/features/newsletters/components/NewsletterDetail';
import { useMobileHeader } from '#/components/MobileHeader';
import { newsletterOptions } from '#/features/newsletters/queries/newsletters';
import { NEWSLETTER_TABS, parseTab } from '#/features/newsletters/lib/tabs';
import type { NewsletterTab } from '#/features/newsletters/lib/tabs';

const NewsletterPage = () => {
  const { newsletterId } = Route.useParams();
  const { tab = 'feeds' } = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: newsletter } = useSuspenseQuery(
    newsletterOptions(newsletterId),
  );

  useMobileHeader({ title: newsletter.name, back: { to: '/newsletters' } });

  return (
    <div className="mx-auto max-w-5xl px-4 py-6 md:px-6 md:py-8 lg:px-10">
      <NewsletterDetail
        newsletter={newsletter}
        tab={tab}
        onTabChange={(next) =>
          navigate({
            search: { tab: next === 'feeds' ? undefined : next },
            replace: true,
            resetScroll: false,
          })
        }
      />
    </div>
  );
};

export const Route = createFileRoute('/newsletters/$newsletterId/')({
  component: NewsletterPage,
  validateSearch: (
    search: Record<string, unknown>,
  ): { tab?: NewsletterTab } => ({
    tab: parseTab(NEWSLETTER_TABS, search.tab),
  }),
});
