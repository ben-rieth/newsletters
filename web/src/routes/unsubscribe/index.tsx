import { createFileRoute } from '@tanstack/react-router';
import UnsubscribePage from '#/features/unsubscribe/components/UnsubscribePage';

const UnsubscribeRoute = () => {
  const { unsubscribeToken } = Route.useSearch();

  return <UnsubscribePage token={unsubscribeToken} />;
};

export const Route = createFileRoute('/unsubscribe/')({
  component: UnsubscribeRoute,
  validateSearch: (search: Record<string, unknown>) => ({
    unsubscribeToken: String(search.unsubscribeToken ?? ''),
  }),
});
