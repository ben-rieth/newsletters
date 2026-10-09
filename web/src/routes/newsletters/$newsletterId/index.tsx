import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { z } from 'zod';
import {
  NEWSLETTER_TABS,
  NewsletterDetail,
} from '#/features/newsletters/components/NewsletterDetail';
import { useMobileHeader } from '#/components/MobileHeader';
import { newsletterOptions } from '#/features/newsletters/queries/newsletters';

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
  validateSearch: z.object({
    tab: z.enum(NEWSLETTER_TABS).optional().catch(undefined),
  }),
});
