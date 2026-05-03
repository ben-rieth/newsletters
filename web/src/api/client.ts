import createClient from 'openapi-fetch';
import type { paths } from './schema';
import { clearSession } from '#/features/auth/lib/session';

const client = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_URL,
  credentials: 'include',
});

// Save a clone of each request before the body is consumed so we can retry
// after a token refresh.
const requestClones = new WeakMap<Request, Request>();

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

    const refreshResponse = await fetch(
      `${import.meta.env.VITE_API_URL}/auth/refresh`,
      { method: 'POST', credentials: 'include' },
    );

    if (!refreshResponse.ok) {
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
