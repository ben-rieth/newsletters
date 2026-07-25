import { createFileRoute, Link } from '@tanstack/react-router';
import { Button } from '#/components/ui/button';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';

const LandingPage = () => {
  const isSignedIn = useIsSignedIn();

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] gap-6 text-center px-4">
      <div className="space-y-3">
        <h1 className="text-4xl font-bold tracking-tight">
          Custom Newsletters
        </h1>
        <p className="text-muted-foreground text-lg max-w-md">
          Build and schedule personalized newsletters from your favorite RSS
          feeds.
        </p>
      </div>
      {isSignedIn ? (
        <Button size="lg">
          <Link to="/issues">Go to Issues</Link>
        </Button>
      ) : (
        <div className="flex gap-3">
          <Button size="lg">
            <Link to="/sign-in">Sign In</Link>
          </Button>
          <Button size="lg" variant="outline">
            <Link to="/sign-up">Sign Up</Link>
          </Button>
        </div>
      )}
    </div>
  );
};

export const Route = createFileRoute('/')({
  component: LandingPage,
});
