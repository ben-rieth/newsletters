import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { H2 } from '#/components/ui/typography';
import { NewsletterDetail } from '#/features/newsletters/components/NewsletterDetail';
import { newsletterOptions } from '#/features/newsletters/queries/newsletters';

const NewsletterPage = () => {
  const { newsletterId } = Route.useParams();
  const { data: newsletter } = useSuspenseQuery(
    newsletterOptions(newsletterId),
  );

  return (
    <div className="p-6 space-y-6">
      <div>
        <Link
          to="/newsletters"
          className="text-sm text-muted-foreground hover:underline"
        >
          ← Newsletters
        </Link>
        <H2 className="mt-2">{newsletter.name}</H2>
      </div>

      <NewsletterDetail newsletter={newsletter} />
    </div>
  );
};

export const Route = createFileRoute('/newsletters/$newsletterId/')({
  component: NewsletterPage,
});
