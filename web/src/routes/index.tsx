import { createFileRoute, Link } from '@tanstack/react-router';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';

const linkClass =
  'underline-offset-4 transition-colors hover:text-foreground hover:underline';

const LandingPage = () => {
  const isSignedIn = useIsSignedIn();

  return (
    <div className="flex min-h-full items-center bg-background px-6 py-16 text-foreground sm:px-16 lg:px-24">
      <div className="max-w-md">
        <h1 className="inline-flex items-center gap-2 text-lg font-semibold tracking-tight">
          <span className="h-4 w-1 rounded-full bg-primary" />
          Slowfeed
        </h1>

        <p className="mt-6 leading-relaxed text-muted-foreground">
          Slowfeed turns RSS and Atom feeds into scheduled email newsletters.
          You group feeds, choose when each newsletter sends, and read the
          issues by email or in the app.
        </p>

        <div className="mt-8">
          <Link
            to={isSignedIn ? '/issues' : '/sign-in'}
            className="font-medium text-foreground underline underline-offset-4"
          >
            {isSignedIn ? 'Go to issues' : 'Sign in'}
          </Link>
        </div>

        <nav className="mt-16 flex gap-6 text-sm text-muted-foreground">
          <Link to="/about" className={linkClass}>
            About
          </Link>
          <a href="https://benriethmeier.dev" className={linkClass}>
            benriethmeier.dev
          </a>
        </nav>
      </div>
    </div>
  );
};

export const Route = createFileRoute('/')({
  component: LandingPage,
});
