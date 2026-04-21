import { Outlet, createRootRouteWithContext } from '@tanstack/react-router';
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools';
import { TanStackDevtools } from '@tanstack/react-devtools';
import { ReactQueryDevtoolsPanel } from '@tanstack/react-query-devtools';

import '../styles.css';
import Header from '#/components/Header';
import DebugPanel from '#/features/debug/components/DebugPanel';
import { useRefreshOnLoad } from '#/features/auth/queries/hooks/useRefreshOnLoad';
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
    <div className="flex min-h-screen flex-col items-center justify-center gap-2 p-6 text-center">
      <p className="text-sm font-medium text-destructive">{message}</p>
    </div>
  );
}

function RootComponent() {
  useRefreshOnLoad();

  return (
    <>
      <Header />
      <main className="px-4 max-w-4xl mx-auto w-full">
        <Outlet />
      </main>
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
