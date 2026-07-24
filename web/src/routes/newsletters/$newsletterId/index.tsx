import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { NewsletterDetail } from '#/features/newsletters/components/NewsletterDetail';
import { newsletterOptions } from '#/features/newsletters/queries/newsletters';

const NewsletterPage = () => {
  const { newsletterId } = Route.useParams();
  const { data: newsletter } = useSuspenseQuery(
    newsletterOptions(newsletterId),
  );

  return (
    <div className="mx-auto max-w-5xl px-6 py-8 lg:px-10">
      <NewsletterDetail newsletter={newsletter} />
    </div>
  );
};

export const Route = createFileRoute('/newsletters/$newsletterId/')({
  component: NewsletterPage,
});
