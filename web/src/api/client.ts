import createClient from 'openapi-fetch';
import type { paths } from './schema';
import { clearSession } from '#/features/auth/lib/session';

const client = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_URL,
});

// Save a clone of each request before the body is consumed so we can retry
// after a token refresh.
const requestClones = new WeakMap<Request, Request>();

client.use({
  onRequest({ request }) {
    if (request.url.includes('/auth/refresh')) {
      return request;
    }

    requestClones.set(request, request.clone());

    const token = localStorage.getItem('token');
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`);
    }
    return request;
  },
  async onResponse({ response, request }) {
    if (response.status !== 401 || request.url.includes('/auth/')) {
      return response;
    }

    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      clearSession();
      return response;
    }

    const refreshResponse = await fetch(
      `${import.meta.env.VITE_API_URL}/auth/refresh`,
      {
        method: 'POST',
        headers: { Authorization: `Bearer ${refreshToken}` },
      },
    );

    if (!refreshResponse.ok) {
      clearSession();
      return response;
    }

    const { tokens } = (await refreshResponse.json()) as {
      tokens: { token: string; refreshToken: string };
    };
    localStorage.setItem('token', tokens.token);
    localStorage.setItem('refreshToken', tokens.refreshToken);

    const saved = requestClones.get(request);
    if (!saved) {
      clearSession();
      return response;
    }

    const headers = new Headers(saved.headers);
    headers.set('Authorization', `Bearer ${tokens.token}`);
    return fetch(new Request(saved, { headers }));
  },
});

export default client;
