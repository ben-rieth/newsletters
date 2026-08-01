import { createRouter } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { routeTree } from './routeTree.gen';
import {
  RouteErrorComponent,
  RouteNotFoundComponent,
} from './components/RouteBoundaries';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // A dropped connection shouldn't surface as an error screen on the
      // first blip; give it a couple of quiet retries first.
      retry: 2,
      staleTime: 30_000,
    },
  },
});

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  scrollRestoration: true,
  defaultErrorComponent: RouteErrorComponent,
  defaultNotFoundComponent: () => <RouteNotFoundComponent />,
  context: {
    queryClient,
  },
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
