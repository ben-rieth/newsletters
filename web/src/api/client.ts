import createClient from 'openapi-fetch';
import type { paths } from './schema';
import { clearSession } from '#/features/auth/lib/session';

const BASE_URL = '/api';

const client = createClient<paths>({
  baseUrl: BASE_URL,
  credentials: 'include',
});

// Save a clone of each request before the body is consumed so we can retry
// after a token refresh.
const requestClones = new WeakMap<Request, Request>();

// Refreshing rotates the token, and a second call presenting the token the server
// just revoked is indistinguishable from a stolen one, so it drops every session
// for the account. A page load firing several requests at once against an expired
// access token would do exactly that, so callers share one in-flight refresh.
let refreshInFlight: Promise<boolean> | null = null;

export const refreshSession = () => {
  refreshInFlight ??= fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  })
    .then((response) => response.ok)
    .catch(() => false)
    .finally(() => {
      refreshInFlight = null;
    });

  return refreshInFlight;
};

client.use({
  onRequest({ request }) {
    if (!request.url.includes('/auth/')) {
      requestClones.set(request, request.clone());
    }
    return request;
  },
  async onResponse({ response, request }) {
    if (response.status !== 401 || request.url.includes('/auth/')) {
      return response;
    }

    const refreshed = await refreshSession();

    if (!refreshed) {
      clearSession();
      return response;
    }

    const saved = requestClones.get(request);
    if (!saved) {
      clearSession();
      return response;
    }

    return fetch(new Request(saved));
  },
});

export default client;
