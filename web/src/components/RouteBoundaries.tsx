import { Link, useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { buttonVariants } from '#/components/ui/button';
import { ErrorState } from '#/components/ErrorState';

const boundaryClass = 'mx-auto max-w-3xl px-4 py-10 md:px-6';

/**
 * Rendered inside the failing route's Outlet, so a bad fetch on one screen
 * leaves the sidebar and tab bar navigable.
 */
export const RouteErrorComponent = ({ error, reset }: ErrorComponentProps) => {
  const router = useRouter();

  return (
    <ErrorState
      className={boundaryClass}
      error={error}
      onRetry={() => {
        reset();
        void router.invalidate();
      }}
    />
  );
};

type NotFoundProps = {
  title?: string;
  description?: string;
};

export const RouteNotFoundComponent = ({
  title = 'Not here',
  description = 'This page doesn’t exist, or it was deleted. Nothing is broken.',
}: NotFoundProps) => (
  <ErrorState
    className={boundaryClass}
    title={title}
    description={description}
    action={
      <Link to="/issues" className={buttonVariants({ variant: 'outline' })}>
        Back to Issues
      </Link>
    }
  />
);
