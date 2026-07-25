import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/newsletters/')({
  beforeLoad: () => {
    throw redirect({ to: '/issues' });
  },
});
