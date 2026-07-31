import {
  Outlet,
  createRootRouteWithContext,
  useRouterState,
} from '@tanstack/react-router';
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools';
import { TanStackDevtools } from '@tanstack/react-devtools';
import { ReactQueryDevtoolsPanel } from '@tanstack/react-query-devtools';

import '../styles.css';
import AppSidebar from '#/components/AppSidebar';
import { ErrorState } from '#/components/ErrorState';
import MobileTabBar from '#/components/MobileTabBar';
import { MobileHeaderProvider } from '#/components/MobileHeader';
import DebugPanel from '#/features/debug/components/DebugPanel';
import { useRefreshOnLoad } from '#/features/auth/queries/hooks/useRefreshOnLoad';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';
import type { QueryClient } from '@tanstack/react-query';

interface RouterContext {
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootComponent,
  errorComponent: RootErrorComponent,
});

function RootErrorComponent({ error }: { error: unknown }) {
  return (
    <main className="min-h-full px-5 py-16 sm:px-6 lg:px-10">
      <ErrorState
        className="mx-auto px-0 md:px-0"
        title="Slowfeed didn’t start"
        description="Something failed before the app could load. Reloading usually clears it."
        error={error}
        onRetry={() => window.location.reload()}
      />
    </main>
  );
}

/**
 * Loaders hold the previous screen while they run, so navigation has no
 * feedback of its own. This is that feedback.
 */
function RouteProgress() {
  const isLoading = useRouterState({ select: (state) => state.isLoading });
  if (!isLoading) return null;

  return (
    <div
      role="status"
      aria-label="Loading"
      className="animate-route-progress pointer-events-none fixed inset-x-0 top-0 z-50 h-0.5 bg-primary"
    />
  );
}

const CHROMELESS_ROUTES = [
  '/',
  '/sign-in',
  '/sign-up',
  '/verify-email',
  '/bad-link',
];

const isChromeless = (pathname: string) =>
  CHROMELESS_ROUTES.includes(pathname) || pathname.startsWith('/unsubscribe');

function RootComponent() {
  useRefreshOnLoad();
  const isSignedIn = useIsSignedIn();
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const showSidebar = isSignedIn && !isChromeless(pathname);

  return (
    <>
      <RouteProgress />
      {showSidebar ? (
        <MobileHeaderProvider>
          <div className="flex h-full flex-col overflow-hidden md:flex-row">
            <AppSidebar />
            <main className="min-h-0 flex-1 overflow-y-auto">
              <Outlet />
            </main>
            <MobileTabBar />
          </div>
        </MobileHeaderProvider>
      ) : (
        <main className="min-h-full pb-safe-b">
          <Outlet />
        </main>
      )}
      {import.meta.env.DEV && <DebugPanel />}
      {import.meta.env.DEV && (
        <TanStackDevtools
          config={{
            position: 'bottom-right',
          }}
          plugins={[
            {
              name: 'TanStack Router',
              render: <TanStackRouterDevtoolsPanel />,
            },
            {
              name: 'Tanstack Query',
              render: <ReactQueryDevtoolsPanel />,
            },
          ]}
        />
      )}
    </>
  );
}
