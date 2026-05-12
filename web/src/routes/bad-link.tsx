import { createFileRoute, Link } from '@tanstack/react-router';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';

export const Route = createFileRoute('/bad-link')({
  component: BadLinkPage,
});

function BadLinkPage() {
  const isSignedIn = useIsSignedIn();

  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 text-center">
      <p className="text-lg font-semibold">Link not found</p>
      <p className="text-sm text-muted-foreground">
        This link is invalid or has expired.
      </p>
      <Link
        to={isSignedIn ? '/issues' : '/sign-in'}
        className="text-sm hover:underline"
      >
        {isSignedIn ? '← Back to issues' : '← Sign in'}
      </Link>
    </div>
  );
}
