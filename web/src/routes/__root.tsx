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
  const message =
    error instanceof Error ? error.message : 'Something went wrong';

  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-2 p-6 text-center">
      <p className="text-sm font-medium text-destructive">{message}</p>
    </div>
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
      {showSidebar ? (
        <MobileHeaderProvider>
          <div className="flex h-dvh flex-col overflow-hidden md:flex-row">
            <AppSidebar />
            <main className="flex-1 overflow-y-auto">
              <Outlet />
            </main>
            <MobileTabBar />
          </div>
        </MobileHeaderProvider>
      ) : (
        <main className="min-h-dvh pb-safe-b">
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
